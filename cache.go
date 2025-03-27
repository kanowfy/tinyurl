package tinyurl

type Cache interface {
	Get(code string) (string, error)
	Set(code string, url string) error
}
