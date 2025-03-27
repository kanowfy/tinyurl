package mocks

import "github.com/stretchr/testify/mock"

type MockDB struct {
	mock.Mock
}

func (d *MockDB) GetUrl(code string) (string, error) {
	args := d.Called(code)
	return args.String(0), args.Error(1)
}

func (d *MockDB) CreateShortUrl(code string, url string) error {
	args := d.Called(code, url)
	return args.Error(0)
}
