package tinyurl

import "errors"

var (
	ErrNotFound = errors.New("code not found")
)

type DB interface {
	GetUrl(code string) (string, error)
	CreateShortUrl(code string, url string) error
}
