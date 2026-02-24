package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// BackendImplementation implements the backend interface and does the actual "hard work"
type BackendImplementation struct{}

func (st BackendImplementation) ReverseString(s string) (string, error) {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	log.Printf("backend: processing strings.Reverse(%q) => %q", s, runes)

	return string(runes), nil
}

func (st BackendImplementation) RepeatString(s string, count int) (string, error) {
	// using std lib function
	res := strings.Repeat(s, count)
	log.Printf("backend: processing strings.Repeat(%q, %d) => %q", s, count, res)

	return res, nil
}

func (st BackendImplementation) TimeToString(t time.Time) (string, error) {
	kitchen := fmt.Sprintf("%s +%ds", t.Format(time.Kitchen), t.Second()) // Kitchen format is defined as "3:04PM"
	log.Printf("backend: formatting %s to %s", t, kitchen)
	return kitchen, nil
}
