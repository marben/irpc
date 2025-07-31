package main

type formattingErr struct {
	formattingError error
	unformattedCode string
}

func (e *formattingErr) Error() string {
	return e.formattingError.Error()
}

func (e *formattingErr) Unwrap() error {
	return e.formattingError
}
