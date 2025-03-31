.PHONY: run
run:
	DEBUG=1 go run ./cmd/web

.PHONY: host
host:
	@docker compose up -d

.PHONY: buildhost
buildhost:
	@docker compose down
	@docker rmi tinyurl-tini
	@docker compose up -d

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

.PHONY: watchcss
watchcss:
	@echo 'Running tailwindcss codegen...'
	@twd -i ./web/static/input.css -o ./web/static/output.css --watch

.PHONY: minifycss 
minifycss:
	@echo 'Running tailwindcss codegen...'
	@twd -i ./web/static/input.css -o ./web/static/output.css --minify
