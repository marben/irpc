- test inline structs (if avail, try even as param)
- write named maps and other types. perhaps there is a common pattern and we can just have named encoder pattern?
- support aliases
- get rid of .codeblock() in encoders, since almost none use it - we can make a separate interface for it
- allow serialization/deserialization of types implementing encoding.BinaryMarshaller and AppendingMarshaller. this is supported by time.Time and few others
    - see other peoples solution for time.Time (embedded pointers etc...)
- implement bitpacking writer. first for slice of bool, eventually perhaps
    for everything? (could use some benchmarking before that though)
- implement and test sending big messages with different endpoint.MaxMsgLen. client should be splitting msgs etc?