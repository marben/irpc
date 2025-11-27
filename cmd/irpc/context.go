package main

var _ Type = contextType{}

type contextType struct {
	ni namedInfo
}

// genDecFunc implements Type.
func (c contextType) genDecFunc(decoderVarName string, q *qualifier) string {
	return ""
}

// genEncFunc implements Type.
func (c contextType) genEncFunc(encoderVarName string, q *qualifier) string {
	return ""
}

func newContextType(ni namedInfo) contextType {
	return contextType{ni: ni}
}

// name implements Type.
func (c contextType) name(q *qualifier) string {
	return q.qualifyNamedInfo(c.ni)
}

// codeblock implements Type.
func (c contextType) codeblocks(q *qualifier) []string {
	return nil
}
