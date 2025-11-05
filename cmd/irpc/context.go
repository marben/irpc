package main

var _ Type = contextType{}

type contextType struct {
	ni namedInfo
}

func newContextType(ni namedInfo) contextType {
	return contextType{ni: ni}
}

// name implements Type.
func (c contextType) name(q *qualifier) string {
	return q.qualifyNamedInfo(c.ni)
}

// codeblock implements Type.
func (c contextType) codeblock(q *qualifier) string {
	return ""
}

// decode implements Type.
func (c contextType) decode(varId string, existingVars varNames, q *qualifier) string {
	return "// no code for context decoding\n"
}

// encode implements Type.
func (c contextType) encode(varId string, existingVars varNames, q *qualifier) string {
	return "// no code for context encoding\n"
}
