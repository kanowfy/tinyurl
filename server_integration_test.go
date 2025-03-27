//go:build integration

package tinyurl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	testredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dbName     = "tinyurl"
	dbUser     = "testuser"
	dbPassword = "testpassword"
)

func setupDB(t testing.TB) (*PostgresDB, *RedisCache, func()) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx, "postgres:14-alpine",
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		postgres.WithDatabase(dbName),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		))
	require.NoError(t, err)

	redisContainer, err := testredis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)

	postgresAddr, err := postgresContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	redisAddr, err := redisContainer.Endpoint(ctx, "")
	require.NoError(t, err)

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s/%s", dbUser, dbPassword, postgresAddr, dbName)
	t.Logf("postgres connection string: %s", dbUrl)

	pool, err := pgxpool.New(ctx, dbUrl)
	require.NoError(t, err)

	stdDB := stdlib.OpenDBFromPool(pool)

	require.NoError(t, goose.Up(stdDB, "migrations"))

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	db := NewPostgresDB(pool)
	cache := NewRedisCache(redisClient)

	cleanup := func() {
		postgresContainer.Terminate(ctx)
		redisContainer.Terminate(ctx)
	}

	return db, cache, cleanup
}

func TestServer_Integration(t *testing.T) {
	db, cache, clean := setupDB(t)
	defer clean()

	srv := NewServer(db, cache)

	longUrl := "https://go.dev/"

	resp := httptest.NewRecorder()
	srv.ServeHTTP(resp, newShortenRequest(t, longUrl))
	assert.Equal(t, http.StatusCreated, resp.Code)
	code := readCodeFromResponse(t, resp)

	resp = httptest.NewRecorder()
	srv.ServeHTTP(resp, newShortenRequest(t, "invalidurl"))
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	resp = httptest.NewRecorder()
	srv.ServeHTTP(resp, newRedirectRequest(t, code))
	assert.Equal(t, http.StatusMovedPermanently, resp.Code)
	assertLocation(t, resp, longUrl)

	resp = httptest.NewRecorder()
	srv.ServeHTTP(resp, newRedirectRequest(t, "invalidcode"))
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func readCodeFromResponse(t testing.TB, resp *httptest.ResponseRecorder) string {
	t.Helper()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(b)
}
