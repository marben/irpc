package main

import (
	"fmt"
	"go/types"
)

func varEncoder(t types.Type) (encoder, error) {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Bool:
			return boolEncoder, nil
		case types.Int:
			return intEncoder, nil
		case types.Uint:
			return uintEncoder, nil
		case types.Int8:
			return int8Encoder, nil
		case types.Uint8: // serves 'types.Byte' as well
			return uint8Encoder, nil
		case types.Int16:
			return int16Encoder, nil
		case types.Uint16:
			return uint16Encoder, nil
		case types.Int32: // serves 'types.Rune' as well
			return int32Encoder, nil
		case types.Uint32:
			return uint32Encoder, nil
		case types.Int64:
			return int64Encoder, nil
		case types.Uint64:
			return uint64Encoder, nil
		case types.Float32:
			return float32Encoder, nil
		case types.Float64:
			return float64Encoder, nil
		case types.String:
			return stringEncoder{}, nil
		default:
			return nil, fmt.Errorf("unsupported basic type '%s'", t.Name())
		}
	case *types.Slice:
		return newSliceEncoder(t)
	case *types.Named:
		name := t.Obj().Name()
		switch ut := t.Underlying().(type) {
		case *types.Struct:
			return newNamedStructEncoder(ut)
		case *types.Interface:
			return newInterfaceEncoder(name, ut)
		default:
			return nil, fmt.Errorf("unsupported named type: %s", ut)
		}
	default:
		return nil, fmt.Errorf("unsupported type '%T' of %s ", t, t)
	}

}
