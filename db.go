package tinyurl

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("code not found")
)

type DB interface {
	GetUrl(code string) (string, error)
	CreateShortUrl(code string, url string) error
	GetCodeIfUrlExists(url string) (string, bool)
}

type PostgresDB struct {
	dbpool *pgxpool.Pool
}

func NewPostgresDB(pool *pgxpool.Pool) *PostgresDB {
	return &PostgresDB{pool}
}

func (p *PostgresDB) GetUrl(code string) (string, error) {
	var url string
	stmt := "SELECT long_url FROM urls WHERE code = $1"

	err := p.dbpool.QueryRow(context.Background(), stmt, code).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		} else {
			return "", fmt.Errorf("querying database: %w", err)
		}
	}

	return url, nil
}

func (p *PostgresDB) CreateShortUrl(code string, url string) error {
	stmt := "INSERT INTO urls (code, long_url) VALUES ($1, $2)"
	_, err := p.dbpool.Exec(context.Background(), stmt, code, url)
	if err != nil {
		return fmt.Errorf("insert into db: %w", err)
	}

	return nil
}

func (p *PostgresDB) GetCodeIfUrlExists(url string) (string, bool) {
	var code string
	stmt := "SELECT code FROM urls WHERE long_url = $1"

	err := p.dbpool.QueryRow(context.Background(), stmt, url).Scan(&code)
	if err != nil {
		return "", false
	}

	return code, true
}
