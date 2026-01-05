include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

# run: run the cmd/web application
.PHONY: run
run:
	go run ./cmd/web -env=development -addr=:4000

## migrations/new name=$1: create a new database migration
.PHONY: migrations/new
migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## migrate/up: apply all up database migrations
.PHONY: migrate/up
migrate/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DEV_DB_DSN} up

## migrate/down: migrate down
.PHONY: migrate/down
migrate/down: confirm
	@echo "Running down migrations..."
	migrate -path ./migrations -database ${DEV_DB_DSN} down

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: tidy and vendor module dependencies and format all .go files
.PHONY: tidy
tidy: 
	@echo 'Tidying module dependencies...'
	go mod tidy
	@echo 'Verifying and vendoring module dependencies'
	go mod verify
	go mod vendor
	@echo 'Formatting .go files...'
	go fmt ./...

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Vetting code...'
	go vet ./...
	go tool staticcheck ./...
	@echo 'Running tests'
	go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build: build the cmd/web application
.PHONY: build
build:
	@echo 'Building cmd/web'
	go build -ldflags='-s' -o=./bin/web ./cmd/web
	GOOS=linux GOARCH=arm64 go build -ldflags="-s" -o=./bin/linux_arm64/web ./cmd/web

# ==================================================================================== #
# PROD
# ==================================================================================== #

## deploy: deply the app to the pi
.PHONY: deploy
deploy:
	@echo 'Deploying to production...'
	rsync -P -e "ssh -p $(production_host_port)" ./bin/linux_arm64/web byryan@$(production_host_ip):/opt/byryan/
	rsync -rP --delete -e "ssh -p $(production_host_port)" ./migrations byryan@$(production_host_ip):/opt/byryan/
	rsync -P -e "ssh -p $(production_host_port)" ./remote/prod/personalsite.service byryan@$(production_host_ip):~
	rsync -P -e "ssh -p $(production_host_port)" ./remote/prod/Caddyfile byryan@$(production_host_ip):~
	ssh -t -p $(production_host_port) byryan@$(production_host_ip) "source /etc/byryanweb/production.env && migrate -path /opt/byryan/migrations -database \$$PROD_DB_DSN up && sudo mv ~/byryanweb.service /etc/systemd/system/ && sudo systemctl enable byryanweb && sudo systemctl restart byryanweb && sudo mv ~/Caddyfile /etc/caddy/ && sudo systemctl reload caddy"
