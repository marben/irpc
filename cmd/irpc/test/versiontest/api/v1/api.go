package api

type Api interface {
	ApiVersion() (string, error)
}
