# Variables
GO_CMD=go
DOCKER_COMPOSE_CMD=docker-compose
DOCKER_COMPOSE_FILE=docker-compose.yml

# Start Docker Compose in detached mode
.PHONY: setup
setup:
	$(DOCKER_COMPOSE_CMD) -f $(DOCKER_COMPOSE_FILE) up -d

# Run the DB migration
.PHONY: migrate
migrate:
	$(GO_CMD) run migrations/main.go

# Run the Go project
.PHONY: run
run:
	$(GO_CMD) run main.go

# Run Go tests
.PHONY: test
test:
	$(GO_CMD) test ./... -v
