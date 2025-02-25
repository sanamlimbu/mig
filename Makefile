PACKAGE=mig
DOCKER_CONTAINER=$(PACKAGE)-db
SERVER=$(CURDIR)/server
WEB=${CURDIR}/web

DB_USER=${PACKAGE}
DB_PASS=devdev
DB_HOST=localhost
DB_PORT=5435
DB_NAME=postgres
DB_CONNECTION_STRING="postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

DOCKERFILE_PATH=./Dockerfile.pg_cron

.PHONY: docker-build
docker-build:
	docker build -t postgres-pg_cron:alpine-pg_cron -f $(DOCKERFILE_PATH) .

.PHONY: docker-create
docker-create: docker-build
	docker run -d -p ${DB_PORT}:5432 --name ${DOCKER_CONTAINER} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_PASSWORD=${DB_PASS} \
		-e POSTGRES_DB=${DB_NAME} \
		postgres-pg_cron:alpine-pg_cron

.PHONY: docker-start
docker-start:
	docker start ${DOCKER_CONTAINER}

.PHONY: docker-stop
docker-stop:
	docker stop ${DOCKER_CONTAINER}

.PHONY: docker-remove
docker-remove:
	docker rm ${DOCKER_CONTAINER}

.PHONY: docker-setup
docker-setup:
	docker exec -it $(DOCKER_CONTAINER) psql -U $(DB_USER) -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'

.PHONY: db-status
db-status:
	cd $(SERVER) && go tool goose postgres $(DB_CONNECTION_STRING) -dir ./migrations status

.PHONY: db-down
db-down:
	cd $(SERVER) && go tool goose postgres "$(DB_CONNECTION_STRING)" -dir ./migrations down

.PHONY: db-migrate
db-migrate:
	cd $(SERVER) && go tool goose postgres "$(DB_CONNECTION_STRING)" -dir ./migrations up

.PHONY: db-prepare
db-prepare: db-down db-migrate

.PHONY: db-seed
db-seed:
	cd $(SERVER) && go run cmd/server/main.go seed

.PHONY: db-reset
db-reset: db-down db-migrate db-seed

.PHONY: go-mod-tidy
go-mod-tidy:
	cd $(SERVER) && go mod tidy

.PHONY: go-errcheck
go-errcheck:
	cd $(SERVER) && errcheck ./...

.PHONY: go-lint
go-lint:
	cd $(SERVER) && golangci-lint run ./...

.PHONY: go-test
go-test:
	cd $(SERVER) && go test ./testings/... -v

.PHONY: go-mod-download
go-mod-download:
	cd $(SERVER) && go mod download

.PHONY: serve
serve:
	cd $(SERVER) && air -c .air.toml

.PHONY: generate
generate:
	cd $(SERVER) && go tool sqlc generate

.PHONY: web-install
web-install:
	cd ${WEB} && npm ci

.PHONY: web-watch
web-watch:
	cd ${WEB} && npm run dev

.PHONY: build-web
build-web:
	@echo "Building React..."
	cd ${WEB} && npm run build
	
.PHONY: build-server
build-server:
	@echo "Building Go..."
	cd ${SERVER} && GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -v -o ../bin/mig-server ./cmd/server/main.go

.PHONY: build
build: build-web build-server