package irpctestpkg

//go:generate go run ../

// comments in type bock are represented differently than in normal interface definition
type (
	// tbInterfaceA deals with string
	tbInterfaceA interface {
		reverse(string) string
	}

	// however this one deals with ints
	// and also has
	//
	//multiple lines
	/* and some other type of comments
	 */
	tvInterfaceB interface {
		add(a, b int) int
	}
)
