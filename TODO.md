- propagate correct error on service not found (and write a test for it)
- test error on calling nonexistent function
- test function with zero return vals
- see other peoples solution for time.Time (embedded pointers etc...)
- implement bitpacking writer. first for slice of bool, eventually perhaps
    for everything? (could use some benchmarking before that though)
- consider moving the irpc package from `irpc/pkg/irpc` to `irpc/irpc`
- implement and test sending big messages with different endpoint.MaxMsgLen. client should be splitting msgs etc?
- allow usage of '*' or '.' to specify more than one input file