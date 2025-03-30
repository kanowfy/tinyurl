package tinyurl

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/kanowfy/tinyurl/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var dummyRateLimiter *DummyRateLimiter

func TestServer(t *testing.T) {
	mockDB := &mocks.MockDB{}
	mockCache := &mocks.MockCache{}

	goUrl := "https://go.dev/"

	mockCache.On("Set", mock.Anything, mock.Anything).Return(nil)
	mockCache.On("Get", mock.Anything).Return("", ErrNotInCache)

	mockDB.On("GetUrl", "abcdef").Return(goUrl, nil)
	mockDB.On("GetUrl", "invalidcode").Return("", ErrNotFound)
	mockDB.On("CreateShortUrl", mock.Anything, goUrl).Return(nil)
	//TODO: test for existing url
	mockDB.On("GetCodeIfUrlExists", mock.Anything).Return("", false)

	srv := NewServer(mockDB, mockCache, dummyRateLimiter)

	resp := httptest.NewRecorder()
	srv.ServeHTTP(resp, newRedirectRequest(t, "abcdef"))
	assert.Equal(t, http.StatusMovedPermanently, resp.Code)
	assertLocation(t, resp, goUrl)

	resp = httptest.NewRecorder()
	srv.ServeHTTP(resp, newRedirectRequest(t, "invalidcode"))
	assert.Equal(t, http.StatusNotFound, resp.Code)

	resp = httptest.NewRecorder()
	srv.ServeHTTP(resp, newShortenRequest(t, goUrl))
	assert.Equal(t, http.StatusCreated, resp.Code)
	mockDB.AssertCalled(t, "CreateShortUrl", mock.Anything, goUrl)
	mockCache.AssertNumberOfCalls(t, "Get", 2)
	mockCache.AssertNumberOfCalls(t, "Set", 2)
	mockCache.AssertCalled(t, "Get", "abcdef")
}

func newShortenRequest(t testing.TB, longUrl string) *http.Request {
	t.Helper()

	formData := url.Values{}
	formData.Set("long_url", longUrl)
	encoded := formData.Encode()

	req, err := http.NewRequest(http.MethodPost, "/shorten", strings.NewReader(encoded))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

func newRedirectRequest(t testing.TB, code string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", code), nil)
	require.NoError(t, err)

	return req
}

func assertLocation(t testing.TB, resp *httptest.ResponseRecorder, longUrl string) {
	t.Helper()

	h := resp.Result().Header
	loc := h.Get("Location")
	if loc != longUrl {
		t.Errorf("wrong redirect location, want %q got %q", longUrl, loc)
	}
}

type DummyRateLimiter struct{}

func (d *DummyRateLimiter) Allow(ip string) bool {
	return true
}

func (d *DummyRateLimiter) Enabled() bool {
	return false
}
