.PHONY: test
test:
	go test ./... -race

.PHONY: test/integration
test/integration:
	go test -tags=integration ./... -race -v

.PHONY: migrate/up
migrate/up:
	@echo 'Running database up migration...'
	@goose postgres "postgres://postgres:postgres@localhost:5432/tinyurl" up -dir=./migrations/

.PHONY: migrate/reset
migrate/reset:
	@echo 'Running reset migration...'
	@goose postgres "postgres://postgres:postgres@localhost:5432/tinyurl" reset -dir=./migrations/