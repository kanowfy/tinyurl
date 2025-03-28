package tinyurl

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"text/template"
)

const (
	codeLength = 6
)

type Server struct {
	http.Handler
	db    DB
	cache Cache
}

func NewServer(db DB, cache Cache) *Server {
	srv := &Server{db: db, cache: cache}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", srv.handleHome)
	mux.HandleFunc("GET /{code}", srv.handleRedirect)
	mux.HandleFunc("POST /shorten", srv.handleShorten)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	srv.Handler = mux

	return srv
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("./web/templates/home.page.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("parse template: %v", err), http.StatusInternalServerError)
		return
	}

	if err := tpl.Execute(w, nil); err != nil {
		http.Error(w, fmt.Sprintf("execute template: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")

	url, err := s.cache.Get(code)
	if err == nil {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
		return
	}

	url, err = s.db.GetUrl(code)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
		} else {
			http.Error(w, fmt.Sprintf("error getting long form url: %v", err), http.StatusInternalServerError)
		}
		return
	}

	s.cache.Set(code, url)

	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	url := r.FormValue("long_url")
	if err := validateUrl(url); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code, err := generateShortCode()
	if err != nil {
		http.Error(w, fmt.Sprintf("error creating shortened url: %v", err), http.StatusInternalServerError)

	}

	if err := s.db.CreateShortUrl(code, url); err != nil {
		http.Error(w, fmt.Sprintf("error saving url: %v", err), http.StatusInternalServerError)
		return
	}

	// newly created url is likely to be accessed soon
	if err := s.cache.Set(code, url); err != nil {
		slog.Error("failed to set to cache", slog.String("error", err.Error()))
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(code))
}

func generateShortCode() (string, error) {
	buf := make([]byte, codeLength)

	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("filling random bytes: %w", err)
	}

	code := base64.URLEncoding.EncodeToString(buf)
	// trim padding characters
	return strings.TrimRight(code, "=")[:codeLength], nil
}

func validateUrl(longUrl string) error {
	if longUrl == "" {
		return fmt.Errorf("missing url")
	}

	parsed, err := url.ParseRequestURI(longUrl)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("missing scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("invalid host")
	}

	return nil
}
