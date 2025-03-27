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

func TestServer(t *testing.T) {
	mockDB := &mocks.MockDB{}

	goUrl := "https://go.dev/"

	mockDB.On("GetUrl", "abcdef").Return("https://go.dev/", nil)
	mockDB.On("GetUrl", "invalidcode").Return("", ErrNotFound)
	mockDB.On("CreateShortUrl", mock.Anything, goUrl).Return(nil)

	srv := NewServer(mockDB)

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
}

func newShortenRequest(t testing.TB, longUrl string) *http.Request {
	t.Helper()

	formData := url.Values{}
	formData.Set("long_url", longUrl)
	encoded := formData.Encode()
	t.Logf("form data: %v", encoded)

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
