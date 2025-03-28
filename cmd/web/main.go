package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/kanowfy/tinyurl"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

const (
	port         = 8080
	redisAddr    = "localhost:6379"
	postgresAddr = "postgres://postgres:postgres@localhost:5432/tinyurl"
)

func main() {
	ctx := context.Background()
	pgpool, err := pgxpool.New(ctx, postgresAddr)
	if err != nil {
		panic(err)
	}

	stdDB := stdlib.OpenDBFromPool(pgpool)

	if err := goose.Up(stdDB, "migrations"); err != nil {
		panic(err)
	}

	rdclient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	srv := tinyurl.NewServer(tinyurl.NewPostgresDB(pgpool), tinyurl.NewRedisCache(rdclient))

	log.Printf("listening on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), srv))
}
