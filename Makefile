# Include variables from .envrc
include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:' 
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api/ -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to database using psql
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/create name=$1: create a new database migration
.PHONY: db/migrations/create
db/migrations/create:
	@echo "Creating migration files for ${name}..."
	migrate create -seq -ext .sql -dir ./migrations/ ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo "Running up migrations..."
	migrate -path ./migrations/ -database ${GREENLIGHT_DB_DSN} up

# ====================================================================================
# QUALITY CONTROL
# ====================================================================================

## tidy: format all .go files, tidy module dependencies, verify them and vendoring them
.PHONY: tidy
tidy: 
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies'
	go mod verify
	go mod vendor

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo "Vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "Running tests..."
	go test -race -vet=off ./...


## remove-vendor: remove vendor folders
.PHONY: remove-vendor
remove-vendor:
	rm -rf vendor/

# ====================================================================================
# BUILD 
# ====================================================================================
## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo "building cmd/api"
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/api ./cmd/api/

# ====================================================================================
# GREENLIGHT PRODUCTION
# ====================================================================================
production_host_ip=152.42.172.170

## production/connect: connect to production server
.PHONY: production/connect
production/connect:
	ssh greenlight@${production_host_ip}

## production/deploy/api: deploy the api to production
.PHONY: production/deploy/api
production/deploy/api:
	rsync -P ./bin/linux_amd64/api greenlight@${production_host_ip}:~
	rsync -rP --delete ./migrations greenlight@${production_host_ip}:~
	ssh -t greenlight@${production_host_ip} 'migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up'

# ====================================================================================
# RUN PROD
# ====================================================================================
## run-me: run the fucking thing
.PHONY: run-me
run-me:
	./bin/api -port=4040 -db-dsn=${GREENLIGHT_DB_DSN}
