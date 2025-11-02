package main

type namedSymetricDirectCallType struct {
	ni       *namedInfo
	typeName string
	enc      encoder
}

func (tr *typeResolver) newNamedSymmetricDirectCallType(encDecFunc string, typeName string, ni *namedInfo) (Type, error) {
	return namedSymetricDirectCallType{
		ni:       ni,
		enc:      newSymmetricDirectCallEncoder(encDecFunc, typeName, ni),
		typeName: typeName,
	}, nil
}

func (t namedSymetricDirectCallType) Name(q *qualifier) string {
	if t.ni == nil {
		return t.typeName
	}

	return q.qualifyNamedInfo(*t.ni)
}

func (t namedSymetricDirectCallType) codeblock(q *qualifier) string {
	return t.enc.codeblock(q)
}

func (t namedSymetricDirectCallType) decode(varId string, existingVars varNames, q *qualifier) string {
	return t.enc.decode(varId, existingVars, q)
}

func (t namedSymetricDirectCallType) encode(varId string, existingVars varNames, q *qualifier) string {
	return t.enc.encode(varId, existingVars, q)
}
