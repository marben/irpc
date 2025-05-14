package irpctestpkg

//go:generate go run ../

// here to test qualifier passing
// exported struct used to cause  `command-line-arguments` as package in vartype
type FileInfo struct {
	FileSize uint64
}

type FileServer interface {
	ListFiles() ([]FileInfo, error)
}
