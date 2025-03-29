package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/caarlos0/env/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/kanowfy/tinyurl"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

type config struct {
	Port      int    `env:"PORT" envDefault:"8080"`
	DbAddr    string `env:"DB_ADDR"`
	CacheAddr string `env:"CACHE_ADDR"`
}

func loadConfig() (config, error) {
	godotenv.Load()
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return config{}, fmt.Errorf("load environment variables: %w", err)
	}

	return cfg, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	pgpool, err := pgxpool.New(ctx, cfg.DbAddr)
	if err != nil {
		panic(err)
	}

	stdDB := stdlib.OpenDBFromPool(pgpool)

	if err := goose.Up(stdDB, "migrations"); err != nil {
		panic(err)
	}

	rdclient := redis.NewClient(&redis.Options{
		Addr: cfg.CacheAddr,
	})

	srv := tinyurl.NewServer(tinyurl.NewPostgresDB(pgpool), tinyurl.NewRedisCache(rdclient))

	log.Printf("listening on port %d\n", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), srv))
}
