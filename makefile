OS := $(shell uname -s 2>/dev/null || echo Windows)

# Variables
ENVIRONMENT ?= dev
CONTAINER_NAME=${APP_NAME}-app-${ENVIRONMENT}
POSTGRES_CONTAINER_NAME=${APP_NAME}-db-${ENVIRONMENT}

init-dev:
	docker-compose -f docker-compose.dev.yml up --build -d

up-dev: 
	docker-compose -f docker-compose.dev.yml up -d