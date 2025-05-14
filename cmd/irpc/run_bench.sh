#!/bin/sh
echo "writing benchmark to bench.txt"
go test -bench=. ./...  -benchmem > bench.txt
echo "done"