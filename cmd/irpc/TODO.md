- dot imports
- BinaryAppender
- test named types outside of module
- support/test aliases
- direct slice of bools optimized encoder in irpc package
- get rid of .codeblock() in encoders, since almost none use it - we can make a separate interface for it
- boolslice and byteslice generators don't really differ much. merge?
- (maybe) implement bitpacking writer. first for slice of bool, eventually perhaps
    for everything? (could use some benchmarking before that though)
- implement and test sending big messages with different endpoint.MaxMsgLen. client should be splitting msgs etc?
