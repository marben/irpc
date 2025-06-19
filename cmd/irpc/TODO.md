- test error on calling nonexistent function
- test function with zero return vals
- see other peoples solution for time.Time (embedded pointers etc...)
- implement bitpacking writer. first for slice of bool, eventually perhaps
    for everything? (could use some benchmarking before that though)
- implement and test sending big messages with different endpoint.MaxMsgLen. client should be splitting msgs etc?