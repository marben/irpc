- get rid of all panic
 - when service is not running on the other end, and our function returns error, return some irpc error
- figure out error handling and what to do on server side panics
- test function with zero return vals
- slice - consider passing capacity
    - consider capacity's max size (and a limit for all calls for that sake)
    - slices (and other reference types) passed in as value can't be changed. i guess, this cannot be avoided
        but has to be mentioned somewhere
- see other peoples solution for time.Time (embedded pointers etc...)
- implement bitpacking writer. first for slice of bool, eventually perhaps
    for everything? (could use some benchmarking before that though)
- check capnproto - something similar?
- get rid of fmt import(mostly used for errors) to possibly reduce size of wasm binary?
- consider moving the irpc package from `irpc/pkg/irpc` to `irpc/irpc`
- implement and test sending big messages with different endpoint.MaxMsgLen. client should be splitting msgs etc?
- allow usage of '*' or '.' to specify more than one input file
- make endpoint.RegisterClient() unexported