package mocks

import "github.com/stretchr/testify/mock"

type MockCache struct {
	mock.Mock
}

func (c *MockCache) Get(code string) (string, error) {
	args := c.Called(code)
	return args.String(0), args.Error(1)
}

func (c *MockCache) Set(code string, url string) error {
	args := c.Called(code)
	return args.Error(0)
}
