package main

import (
	"fmt"
	"go/types"
	"strings"
)

// streamMethodGenerator generates a method with channels (streams).
//
// The user's method is never called over the network directly. It is split into "wire" methods,
// generated like any other method (req/resp structs, dispatch, client call):
//
//	streamOpen_M(ctx, session) error                          registers the session
//	stream_M(ctx, session, scalar params...) (...)            main call, lives as long as the session
//	streamPush_M_ch(ctx, session, values []T) error           for each channel parameter
//	streamClose_M_ch(ctx, session) error                      for each channel parameter
//	streamReady_M(ctx, session) (scalar results..., error)    only if the method returns channels
//	streamPull_M_ch(ctx, session) ([]T, bool, error)          for each returned channel
//
// The main call returns the user's results when the method has no returned channels.
// Otherwise it returns only an error, and the results are collected by Ready.
//
// The service side is implemented by the generated "streams" struct, the client side by the
// generated client method. Both are built on the helpers of irpcgen/stream.go.
type streamMethodGenerator struct {
	apiName string
	user    methodGenerator // the method as declared by the user (with channels)

	ctxParam      genParam
	scalarParams  []genParam // parameters, except context and channels
	inChans       []genParam // channel parameters (caller -> implementation)
	scalarResults []genParam // results, except channels and the final error
	outChans      []genParam // returned channels (implementation -> caller)

	open  methodGenerator
	main  methodGenerator
	ready methodGenerator // only if len(outChans) > 0
	push  []methodGenerator
	close []methodGenerator
	pull  []methodGenerator
}

// validateStreamMethod checks the signature of a method with channels and returns a clear error otherwise.
func validateStreamMethod(tr typeResolver, m methodGenerator) error {
	q := newQualifier(tr)
	for _, p := range m.req.params {
		ct, ok := p.typ.(chanType)
		if !ok {
			continue
		}
		switch ct.dir {
		case types.SendOnly:
			return fmt.Errorf("parameter %q has type %s: only receive-only channels (<-chan T) are supported. "+
				"A channel parameter is a stream from the caller to the implementation, so the implementation can only receive from it",
				p.identifier, ct.name(q))
		case types.SendRecv:
			return fmt.Errorf("parameter %q has type %s: bidirectional channels are not supported, use <-chan %s "+
				"(a stream from the caller to the implementation)",
				p.identifier, ct.name(q), ct.elemT.name(q))
		}
	}

	for i, p := range m.resp.params {
		ct, ok := p.typ.(chanType)
		if !ok {
			continue
		}
		switch ct.dir {
		case types.SendOnly:
			return fmt.Errorf("result %d has type %s: only receive-only channels (<-chan T) can be returned. "+
				"A returned channel is a stream from the implementation to the caller, so the caller can only receive from it",
				i, ct.name(q))
		case types.SendRecv:
			return fmt.Errorf("result %d has type %s: bidirectional channels are not supported, return <-chan %s "+
				"(a stream from the implementation to the caller)",
				i, ct.name(q), ct.elemT.name(q))
		}
	}

	hasCtx := false
	for _, p := range m.req.params {
		if p.isContext() {
			hasCtx = true
		}
	}
	if !hasCtx {
		return fmt.Errorf("methods with channels must have a context.Context parameter: " +
			"it is the context of the streams, and the caller cancels it to stop them")
	}

	if !m.resp.isLastTypeError(q) {
		return fmt.Errorf("methods with channels must return error as the last result: " +
			"streams can fail on the network, and the generated client must be able to report it")
	}

	return nil
}

func newStreamMethodGenerator(tr typeResolver, apiName string, user methodGenerator, nextAuxIndex *int) (streamMethodGenerator, error) {
	if err := validateStreamMethod(tr, user); err != nil {
		return streamMethodGenerator{}, err
	}

	sm := streamMethodGenerator{
		apiName: apiName,
		user:    user,
	}

	for _, p := range user.req.params {
		switch {
		case p.isContext():
			sm.ctxParam = p
		case isChanType(p.typ):
			sm.inChans = append(sm.inChans, p)
		default:
			sm.scalarParams = append(sm.scalarParams, p)
		}
	}
	results := user.resp.params
	for _, p := range results[:len(results)-1] { // the last one is the error
		if isChanType(p.typ) {
			sm.outChans = append(sm.outChans, p)
		} else {
			sm.scalarResults = append(sm.scalarResults, p)
		}
	}
	errParam := results[len(results)-1]

	sessionT := newDirectCallType("irpcgen.EncUint64", "irpcgen.DecUint64", "uint64", nil)
	ctxP := rpcParam{name: "irpcCtx", typ: sm.ctxParam.typ}
	sessionP := rpcParam{name: "irpcSession", typ: sessionT}
	errR := rpcParam{typ: errParam.typ}

	auxIndex := func() int {
		i := *nextAuxIndex
		*nextAuxIndex++
		return i
	}

	var err error
	name := user.name

	// open
	sm.open, err = newWireMethod(apiName, "streamOpen_"+name, auxIndex(), []rpcParam{ctxP, sessionP}, []rpcParam{errR}, false)
	if err != nil {
		return streamMethodGenerator{}, err
	}

	// main: keeps the index of the user's method
	mainParams := []rpcParam{ctxP, sessionP}
	for _, p := range sm.scalarParams {
		mainParams = append(mainParams, rpcParam{name: p.identifier, typ: p.typ})
	}
	var mainResults []rpcParam
	if len(sm.outChans) == 0 {
		for _, p := range results {
			mainResults = append(mainResults, rpcParam{typ: p.typ})
		}
	} else {
		mainResults = []rpcParam{errR}
	}
	sm.main, err = newWireMethod(apiName, "stream_"+name, user.index, mainParams, mainResults, true)
	if err != nil {
		return streamMethodGenerator{}, err
	}

	// push/close for each channel parameter
	for _, p := range sm.inChans {
		batchT, err := tr.batchType(apiName, p.typ.(chanType))
		if err != nil {
			return streamMethodGenerator{}, fmt.Errorf("parameter %q: %w", p.identifier, err)
		}
		push, err := newWireMethod(apiName, sm.pushName(p), auxIndex(),
			[]rpcParam{ctxP, sessionP, {name: "irpcValues", typ: batchT}}, []rpcParam{errR}, true)
		if err != nil {
			return streamMethodGenerator{}, err
		}
		cls, err := newWireMethod(apiName, sm.closeName(p), auxIndex(), []rpcParam{ctxP, sessionP}, []rpcParam{errR}, false)
		if err != nil {
			return streamMethodGenerator{}, err
		}
		sm.push = append(sm.push, push)
		sm.close = append(sm.close, cls)
	}

	// ready + pull for each returned channel
	if len(sm.outChans) > 0 {
		var readyResults []rpcParam
		for _, p := range sm.scalarResults {
			readyResults = append(readyResults, rpcParam{typ: p.typ})
		}
		readyResults = append(readyResults, errR)
		sm.ready, err = newWireMethod(apiName, "streamReady_"+name, auxIndex(), []rpcParam{ctxP, sessionP}, readyResults, true)
		if err != nil {
			return streamMethodGenerator{}, err
		}

		for i, p := range sm.outChans {
			batchT, err := tr.batchType(apiName, p.typ.(chanType))
			if err != nil {
				return streamMethodGenerator{}, fmt.Errorf("result %d: %w", i, err)
			}
			pull, err := newWireMethod(apiName, sm.pullName(p), auxIndex(), []rpcParam{ctxP, sessionP},
				[]rpcParam{{typ: batchT}, {typ: tr.boolType}, errR}, true)
			if err != nil {
				return streamMethodGenerator{}, err
			}
			sm.pull = append(sm.pull, pull)
		}
	}

	return sm, nil
}

func newWireMethod(apiName, name string, index int, params, results []rpcParam, unlimited bool) (methodGenerator, error) {
	for i := range params {
		params[i].pos = i
	}
	for i := range results {
		results[i].pos = i
	}
	req, resp, err := newReqRespStructsGenerator(apiName, name, params, results)
	if err != nil {
		return methodGenerator{}, fmt.Errorf("newReqRespStructsGenerator(%s): %w", name, err)
	}
	return methodGenerator{
		name:      name,
		index:     index,
		req:       req,
		resp:      resp,
		ctxVar:    "irpcCtx",
		receiver:  "streams",
		unlimited: unlimited,
		isWire:    true,
	}, nil
}

func (sm streamMethodGenerator) pushName(p genParam) string {
	return "streamPush_" + sm.user.name + "_" + p.identifier
}
func (sm streamMethodGenerator) closeName(p genParam) string {
	return "streamClose_" + sm.user.name + "_" + p.identifier
}
func (sm streamMethodGenerator) pullName(p genParam) string {
	return "streamPull_" + sm.user.name + "_" + p.structFieldName
}

// wireMethods returns all the wire methods, which are generated as normal methods.
func (sm streamMethodGenerator) wireMethods() []methodGenerator {
	ms := []methodGenerator{sm.open, sm.main}
	for i := range sm.push {
		ms = append(ms, sm.push[i], sm.close[i])
	}
	if len(sm.outChans) > 0 {
		ms = append(ms, sm.ready)
		ms = append(ms, sm.pull...)
	}
	return ms
}

func (sm streamMethodGenerator) hasOutChans() bool {
	return len(sm.outChans) > 0
}

func (sm streamMethodGenerator) sessionTypeName() string {
	return "_irpc_" + sm.apiName + "_" + sm.user.name + "Session"
}

func (sm streamMethodGenerator) sessionsFieldName() string {
	return "sessions_" + sm.user.name
}

func inFieldName(p genParam) string     { return "in_" + p.identifier }
func outFieldName(p genParam) string    { return "out_" + p.structFieldName }
func resultFieldName(p genParam) string { return "r_" + p.structFieldName }

// zeroReturn returns "var irpcZero0 T0 ...; return irpcZero0, ..., errExpr"
// results must end with the error, which is replaced by errExpr
func zeroReturn(q *qualifier, results []genParam, errExpr string) string {
	sb := &strings.Builder{}
	vals := []string{}
	for i, p := range results[:len(results)-1] {
		z := fmt.Sprintf("irpcZero%d", i)
		fmt.Fprintf(sb, "var %s %s\n", z, p.typ.name(q))
		vals = append(vals, z)
	}
	vals = append(vals, errExpr)
	fmt.Fprintf(sb, "return %s\n", strings.Join(vals, ", "))
	return sb.String()
}

// wireSignature returns "name(params) (results)" of a wire method, as implemented by the streams struct
func wireSignature(q *qualifier, m methodGenerator) string {
	return fmt.Sprintf("%s(%s) (%s)", m.name, m.req.funcCallParams(q), m.resp.funcCallParams(q))
}

// implArgs returns the arguments to call the user's implementation from the main call
func (sm streamMethodGenerator) implArgs() string {
	args := []string{}
	for _, p := range sm.user.req.params {
		switch {
		case p.isContext():
			args = append(args, "irpcSess.Context()")
		case isChanType(p.typ):
			args = append(args, "irpcSess."+inFieldName(p)+".Chan()")
		default:
			args = append(args, p.identifier)
		}
	}
	return strings.Join(args, ", ")
}

// getSession is the common start of the service side wire methods
func (sm streamMethodGenerator) getSession(q *qualifier, m methodGenerator) string {
	return fmt.Sprintf(`irpcSess, irpcErr := irpcS.%s.Get(irpcCtx, irpcSession)
		if irpcErr != nil {
			%s}
		`, sm.sessionsFieldName(), zeroReturn(q, m.resp.params, "irpcErr"))
}

// enterSession is getSession for the calls that can't overlap on the same key (see irpcgen.StreamSession.Enter)
func (sm streamMethodGenerator) enterSession(q *qualifier, m methodGenerator, key string) string {
	return sm.getSession(q, m) + fmt.Sprintf(`irpcExit, irpcErr := irpcSess.Enter(%q)
		if irpcErr != nil {
			%s}
		defer irpcExit()
		`, key, zeroReturn(q, m.resp.params, "irpcErr"))
}

// serverCode generates the session type and the service side of the wire methods
func (sm streamMethodGenerator) serverCode(q *qualifier, streamsType string) string {
	sb := &strings.Builder{}

	// session type
	fmt.Fprintf(sb, "// %s is the service side state of one [%s.%s] call\n", sm.sessionTypeName(), sm.apiName, sm.user.name)
	fmt.Fprintf(sb, "type %s struct {\n*irpcgen.StreamSession\n", sm.sessionTypeName())
	for _, p := range sm.inChans {
		fmt.Fprintf(sb, "%s *irpcgen.InStream[%s]\n", inFieldName(p), p.typ.(chanType).elemT.name(q))
	}
	if sm.hasOutChans() {
		for _, p := range sm.outChans {
			fmt.Fprintf(sb, "%s *irpcgen.OutStream[%s]\n", outFieldName(p), p.typ.(chanType).elemT.name(q))
		}
		for _, p := range sm.scalarResults {
			fmt.Fprintf(sb, "%s %s\n", resultFieldName(p), p.typ.name(q))
		}
		sb.WriteString("implErr error\n")
	}
	sb.WriteString("}\n\n")

	recv := fmt.Sprintf("func (irpcS *%s) ", streamsType)

	// open
	fmt.Fprintf(sb, "%s%s {\n", recv, wireSignature(q, sm.open))
	fmt.Fprintf(sb, "return irpcS.%s.Add(irpcCtx, irpcSession, &%s{\nStreamSession: irpcgen.NewStreamSession(irpcCtx),\n", sm.sessionsFieldName(), sm.sessionTypeName())
	for _, p := range sm.inChans {
		fmt.Fprintf(sb, "%s: irpcgen.NewInStream[%s](),\n", inFieldName(p), p.typ.(chanType).elemT.name(q))
	}
	sb.WriteString("})\n}\n\n")

	// main
	fmt.Fprintf(sb, "%s%s {\n", recv, wireSignature(q, sm.main))
	sb.WriteString(sm.getSession(q, sm.main))
	// before the deferred Del: a rejected main call must not end the session of the one running
	fmt.Fprintf(sb, `if irpcErr := irpcSess.Start(); irpcErr != nil {
			%s}
		`, zeroReturn(q, sm.main.resp.params, "irpcErr"))
	fmt.Fprintf(sb, `irpcStop := irpcSess.Link(irpcCtx)
		defer func() {
			irpcStop()
			irpcSess.End(irpcgen.ErrStreamSessionEnded)
			irpcS.%s.Del(irpcCtx, irpcSession)
		}()
		`, sm.sessionsFieldName())
	if !sm.hasOutChans() {
		// the implementation reads until it returns: then the session ends
		fmt.Fprintf(sb, "return irpcS.impl.%s(%s)\n}\n\n", sm.user.name, sm.implArgs())
	} else {
		targets := []string{}
		for _, p := range sm.user.resp.params[:len(sm.user.resp.params)-1] {
			if isChanType(p.typ) {
				local := "irpcOut_" + p.structFieldName
				fmt.Fprintf(sb, "var %s %s\n", local, p.typ.name(q))
				targets = append(targets, local)
			} else {
				targets = append(targets, "irpcSess."+resultFieldName(p))
			}
		}
		targets = append(targets, "irpcSess.implErr")
		fmt.Fprintf(sb, "%s = irpcS.impl.%s(%s)\n", strings.Join(targets, ", "), sm.user.name, sm.implArgs())
		sb.WriteString("if irpcSess.implErr == nil {\n")
		drained := []string{}
		for _, p := range sm.outChans {
			fmt.Fprintf(sb, "irpcSess.%s = irpcgen.NewOutStream(irpcOut_%s)\n", outFieldName(p), p.structFieldName)
			drained = append(drained, "irpcSess."+outFieldName(p)+".Drained()")
		}
		sb.WriteString("}\n")
		fmt.Fprintf(sb, `irpcSess.MarkReady()
			if irpcSess.implErr != nil {
				// keep the session until Ready has collected the error
				irpcSess.WaitReadyTaken()
				return irpcSess.implErr
			}
			// the session lives until the caller has received everything
			return irpcSess.WaitDrained(%s)
		}

		`, strings.Join(drained, ", "))
	}

	// push/close
	for i, p := range sm.inChans {
		fmt.Fprintf(sb, "%s%s {\n", recv, wireSignature(q, sm.push[i]))
		sb.WriteString(sm.enterSession(q, sm.push[i], inFieldName(p)))
		fmt.Fprintf(sb, "return irpcSess.%s.Deliver(irpcSess.Context(), irpcValues)\n}\n\n", inFieldName(p))

		fmt.Fprintf(sb, "%s%s {\n", recv, wireSignature(q, sm.close[i]))
		sb.WriteString(sm.enterSession(q, sm.close[i], inFieldName(p)))
		fmt.Fprintf(sb, "irpcSess.%s.Close()\nreturn nil\n}\n\n", inFieldName(p))
	}

	// ready/pull
	if sm.hasOutChans() {
		fmt.Fprintf(sb, "%s%s {\n", recv, wireSignature(q, sm.ready))
		sb.WriteString(sm.enterSession(q, sm.ready, "ready"))
		fmt.Fprintf(sb, "if irpcErr := irpcSess.TakeReady(); irpcErr != nil {\n%s}\n", zeroReturn(q, sm.ready.resp.params, "irpcErr"))
		vals := []string{}
		for _, p := range sm.scalarResults {
			vals = append(vals, "irpcSess."+resultFieldName(p))
		}
		vals = append(vals, "irpcSess.implErr")
		fmt.Fprintf(sb, "return %s\n}\n\n", strings.Join(vals, ", "))

		for i, p := range sm.outChans {
			fmt.Fprintf(sb, "%s%s {\n", recv, wireSignature(q, sm.pull[i]))
			sb.WriteString(sm.enterSession(q, sm.pull[i], outFieldName(p)))
			fmt.Fprintf(sb, `if irpcErr := irpcSess.WaitReady(); irpcErr != nil {
					return nil, false, irpcErr
				}
				if irpcSess.implErr != nil {
					return nil, false, irpcSess.implErr
				}
				return irpcSess.%s.Pull(irpcSess.Context())
			}

			`, outFieldName(p))
		}
	}

	return sb.String()
}

// clientCode generates the client's implementation of the user's method, on top of the wire methods
func (sm streamMethodGenerator) clientCode(q *qualifier, clientType string) string {
	sb := &strings.Builder{}
	ctx := sm.ctxParam.identifier
	results := sm.user.resp.params

	fmt.Fprintf(sb, "// %s implements [%s]\n//\n", sm.user.name, sm.apiName)
	sb.WriteString(sm.user.goDoc)
	if sm.hasOutChans() {
		fmt.Fprintf(sb, "//\n// The returned channels are closed when the implementation closes them, or when %s is done: cancel %s to stop reading early.\n", ctx, ctx)
	}
	fmt.Fprintf(sb, "func (_c *%s) %s(%s) (%s) {\n", clientType, sm.user.name, sm.user.req.funcCallParams(q), sm.user.resp.funcCallParams(q))

	fmt.Fprintf(sb, `irpcSession := irpcgen.NewStreamSessionId()
		if irpcErr := _c.%s(%s, irpcSession); irpcErr != nil {
			%s}
		`, sm.open.name, ctx, zeroReturn(q, results, "irpcErr"))

	scalarArgs := ""
	for _, p := range sm.scalarParams {
		scalarArgs += ", " + p.identifier
	}

	// pumps of the channel parameters
	pumpsCtx := "irpcPumpsCtx"
	if sm.hasOutChans() {
		fmt.Fprintf(sb, `irpcSessCtx, irpcCancelSess := context.WithCancelCause(%s)
			irpcMainErrC := make(chan error, 1)
			go func() {
				irpcMainErrC <- _c.%s(irpcSessCtx, irpcSession%s)
			}()
			`, ctx, sm.main.name, scalarArgs)
		pumpsCtx = "irpcSessCtx"
	} else if len(sm.inChans) > 0 {
		fmt.Fprintf(sb, `irpcPumpsCtx, irpcCancelPumps := context.WithCancel(%s)
			defer irpcCancelPumps()
			`, ctx)
	}
	waitPumps := ""
	for i, p := range sm.inChans {
		elem := p.typ.(chanType).elemT.name(q)
		fmt.Fprintf(sb, `irpcWaitPump_%[1]s := irpcgen.GoPumpStream(%[2]s, %[1]s,
				func(irpcCtx context.Context, irpcValues []%[3]s) error { return _c.%[4]s(irpcCtx, irpcSession, irpcValues) },
				func(irpcCtx context.Context) error { return _c.%[5]s(irpcCtx, irpcSession) })
			`, p.identifier, pumpsCtx, elem, sm.push[i].name, sm.close[i].name)
		waitPumps += fmt.Sprintf("irpcWaitPump_%s()\n", p.identifier)
	}

	if !sm.hasOutChans() {
		vals := []string{}
		for i := range results {
			vals = append(vals, fmt.Sprintf("irpcR%d", i))
		}
		fmt.Fprintf(sb, "%s := _c.%s(%s, irpcSession%s)\n", strings.Join(vals, ", "), sm.main.name, ctx, scalarArgs)
		if len(sm.inChans) > 0 {
			sb.WriteString("irpcCancelPumps()\n")
			sb.WriteString(waitPumps)
		}
		fmt.Fprintf(sb, "return %s\n}\n", strings.Join(vals, ", "))
		return sb.String()
	}

	// ready
	readyVals := []string{}
	for _, p := range sm.scalarResults {
		readyVals = append(readyVals, "irpcR_"+p.structFieldName)
	}
	readyVals = append(readyVals, "irpcErr")
	fmt.Fprintf(sb, `%s := _c.%s(irpcSessCtx, irpcSession)
		if irpcErr != nil {
			irpcCancelSess(irpcErr)
			<-irpcMainErrC
			%s%s}
		`, strings.Join(readyVals, ", "), sm.ready.name, waitPumps, zeroReturn(q, results, "irpcErr"))

	// receivers of the returned channels
	for i, p := range sm.outChans {
		elem := p.typ.(chanType).elemT.name(q)
		fmt.Fprintf(sb, `irpcOut_%[1]s := make(chan %[2]s)
			irpcWaitRecv_%[1]s := irpcgen.GoReceiveStream(irpcSessCtx, irpcOut_%[1]s, func(irpcCtx context.Context) ([]%[2]s, bool, error) {
				return _c.%[3]s(irpcCtx, irpcSession)
			})
			`, p.structFieldName, elem, sm.pull[i].name)
	}

	// once everything has been received (or failed), wait for the main call and release the session
	sb.WriteString("go func() {\n")
	for _, p := range sm.outChans {
		fmt.Fprintf(sb, "if irpcErr := irpcWaitRecv_%s(); irpcErr != nil {\nirpcCancelSess(irpcErr)\n}\n", p.structFieldName)
	}
	sb.WriteString("<-irpcMainErrC\nirpcCancelSess(nil)\n")
	sb.WriteString(waitPumps)
	sb.WriteString("}()\n")

	vals := []string{}
	for _, p := range results[:len(results)-1] {
		if isChanType(p.typ) {
			vals = append(vals, "irpcOut_"+p.structFieldName)
		} else {
			vals = append(vals, "irpcR_"+p.structFieldName)
		}
	}
	vals = append(vals, "nil")
	fmt.Fprintf(sb, "return %s\n}\n", strings.Join(vals, ", "))

	return sb.String()
}
