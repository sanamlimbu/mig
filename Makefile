PACKAGE=mig
DOCKER_CONTAINER=$(PACKAGE)-db
SERVER=$(CURDIR)/server
WEB=${CURDIR}/web

DB_USER=${PACKAGE}
DB_PASS=devdev
DB_HOST=localhost
DB_PORT=5435
DB_NAME=$(PACKAGE)
DB_CONNECTION_STRING="postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

.PHONY: docker-create
docker-create:
	docker run -d -p ${DB_PORT}:5432 --name ${DOCKER_CONTAINER} \
		-e POSTGRES_USER=${DB_USER} \
		-e POSTGRES_PASSWORD=${DB_PASS} \
		-e POSTGRES_DB=${DB_NAME} \
		postgres:alpine

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
	goose postgres $(DB_CONNECTION_STRING) -dir $(SERVER)/migrations status

.PHONY: db-down
db-down:
	goose postgres "$(DB_CONNECTION_STRING)" -dir $(SERVER)/migrations down

.PHONY: db-migrate
db-migrate:
	goose postgres "$(DB_CONNECTION_STRING)" -dir $(SERVER)/migrations up

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
	cd $(SERVER) && sqlc generate

.PHONY: web-install
web-install:
	cd ${WEB} && npm ci

.PHONY: web-watch
web-watch:
	cd ${WEB} && npm run dev