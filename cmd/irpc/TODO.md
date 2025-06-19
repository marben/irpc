- test function with zero return vals
- allow serialization/deserialization of types implementing encoding.BinaryMarshaller. this is supporte by time.Time
    - see other peoples solution for time.Time (embedded pointers etc...)
- implement bitpacking writer. first for slice of bool, eventually perhaps
    for everything? (could use some benchmarking before that though)
- implement and test sending big messages with different endpoint.MaxMsgLen. client should be splitting msgs etc?