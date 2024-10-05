PACKAGE=mig

DOCKER_CONTAINER=$(PACKAGE)-db

SERVER=$(CURDIR)/server

BIN=$(CURDIR)/bin

DB_USER=postgres
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
	docker exec -it $(DOCKER_CONTAINER) psql -U $(DB_USER) -c 'CREATE EXTENSION IF NOT EXISTS pg_trgm; CREATE EXTENSION IF NOT EXISTS pgcrypto; CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'

.PHONY: db-drop
db-drop:
	bin/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations drop -f

.PHONY: db-down
db-down:
	bin/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations down

.PHONY: db-migrate
db-migrate:
	bin/migrate -database $(DB_CONNECTION_STRING) -path $(SERVER)/migrations up

.PHONY: db-prepare
db-prepare: db-drop db-migrate

.PHONY: db-seed
db-seed:
	cd server && go run cmd/main.go db --seed

.PHONY: db-reset
db-reset: db-drop db-migrate db-seed

.PHONY: go-mod-tidy
go-mod-tidy:
	cd server && go mod tidy

.PHONY: go-mod-download
go-mod-download:
	cd server && go mod download

.PHONY: dev-tools
dev-tools: go-mod-tidy
	@mkdir -p $(BIN) 
	cd $(SERVER) && go generate -tags tools ./tools/...

.PHONY: serve
serve:
	cd $(SERVER) && ../bin/air -c .air.toml

.PHONY: generate
generate:
	$(BIN)/sqlboiler ${BIN}/sqlboiler-psql --wipe --config $(SERVER)/sqlboiler.toml --output $(SERVER)/models

.PHONY: web-install
web-install:
	cd web/ && npm install

.PHONY: web-watch
web-watch:
	cd web/ && npm run dev