package tcp_example

import "time"

//go:generate go run ../../cmd/irpc

// For external projects, use: go:generate go run github.com/marben/irpc/cmd/irpc@latest $GOFILE

// Backend interface defines functions, that will be avalable to be called over the network
// we return error from every function, because network errors can occur even for functions, that cannot error otherwise
// string utils obviously make no sense to be called remotely, but it's just an example
type Backend interface {
	ReverseString(in string) (string, error)
	RepeatString(in string, n int) (string, error)
	TimeToString(t time.Time) (string, error)
}
