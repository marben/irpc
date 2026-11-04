module github.com/marben/irpc/examples/bidirectional

go 1.24.0

require (
	github.com/brianvoe/gofakeit/v7 v7.18.1
	github.com/coder/websocket v1.8.15
	github.com/marben/irpc v0.0.0
)

require (
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
)

replace github.com/marben/irpc => ../..

tool github.com/marben/irpc/cmd/irpc
