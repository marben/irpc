package single

//go:generate irpc $GOFILE

type Sample interface {
	Greeting(name string) string
}
