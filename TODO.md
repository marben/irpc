# TODO

### Code Generation
- Make generators report their required imports and populate the qualifier list only from parameter structs (not from generated functions), if possible.
- Add a helper that batches serialization, deserialization, and error handling, and call it from generated code to eliminate repetitive error checks.
- Reduce repetitive code in struct serializers (e.g., `vect3x3` in `test_struct_irpc.go`).  
  Consider allowing named serialization functions for sub-structures (e.g., `vect3`).
- Consider encoding `isNil` for pointer-based `binaryMarshaller` types.  
  Might be feasible via generics with separate functions for value and pointer types.
- Remove unnecessary `var zero` declarations when the zero value is obvious (`nil` for interface/slice/map, `""` for string, etc.).
- Implement optional bit-packing support.  
  Start with slices of `bool`; benchmark before extending to other types.
- Add support for dot imports.
- Add the `BinaryAppender`.

### Types & Interfaces
- Verify behavior when passing `interface{}` values that are non-nil but may contain nil pointers; ensure the encoder returns a meaningful non-nil representation.
- Clarify handling of `map[int]interface{}` (no code block needed for the literal `interface{}`).
- Test `[]time.Time` — especially the new slice implementation and the asymmetric encoder.
- Add tests for named types defined outside the module.
- Add support and tests for aliases.

### Module & File Handling
- Add support for files outside the module (currently unsupported and causes failures).
- When two interfaces are defined in one file, confirm that service and client hashes differ if needed.  
  Consider including the interface name in the hash.

### Formatting & Tooling
- Investigate why formatting via the Go stdlib produces different (and arguably worse) output than `gofmt`, especially with nested function definitions (e.g., in struct encoders).  
  - Determine whether a known workaround or fix exists.
- Improve fuzzy testing for `Enc`/`Dec` functions to verify encoding stability.
- Improve tests that rely on “waits” or manual timing.  
  Newer Go test utilities offer controllable time flow; apply them where relevant.

### Documentation
- Document versioning strategy in the README.

### Message Handling
- Implement and test sending large messages with varying `endpoint.MaxMsgLen`.  
  The client should split messages as needed.
