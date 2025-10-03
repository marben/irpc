package main

type byteSliceType struct {
	ni  *namedInfo
	enc encoder
}

var _ Type = byteSliceType{}

func (tr *typeResolver) newByteSliceType(ni *namedInfo) (byteSliceType, error) {
	return byteSliceType{
		ni:  ni,
		enc: newSymmetricDirectCallEncoder("ByteSlice", "[]byte", ni),
	}, nil
}

func (b byteSliceType) Name(q *qualifier) string {
	if b.ni == nil {
		return "[]byte"
	}
	return q.qualifyNamedInfo(*b.ni)
}

func (b byteSliceType) codeblock(q *qualifier) string {
	return b.enc.codeblock(q)
}

func (b byteSliceType) decode(varId string, existingVars varNames, q *qualifier) string {
	return b.enc.decode(varId, existingVars, q)
}

func (b byteSliceType) encode(varId string, existingVars varNames, q *qualifier) string {
	return b.enc.encode(varId, existingVars, q)
}
