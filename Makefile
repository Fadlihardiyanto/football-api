# Project-level Makefile

# === Environment ===
PROJECT_NAME = football-api

# === Docker Commands ===

## Build all Docker images
build:
	docker-compose build

## Start all services in background
up:
	docker-compose up -d

## Stop all services
down:
	docker-compose down

## Restart services
restart:
	docker-compose down && docker-compose up -d

## View logs
logs:
	docker-compose logs -f

## See container status
ps:
	docker-compose ps

## Remove all containers, volumes & networks
clean:
	docker-compose down -v --remove-orphans

# === Golang Commands (Local Dev Only) ===

## Run Web service 
run-web:
	docker-compose up web

## Run Worker
run-worker:
	docker-compose up worker

## Run migration 
migrate-up:
	docker-compose exec web migrate \
	-path=./db/migrations \
	-database="postgres://ayo_indonesia_football:football123@postgres:5432/ayo_indonesia_football?sslmode=disable" \
	up


migrate-down:
	docker-compose exec web migrate \
	-path=./db/migrations \
	-database="postgres://ayo_indonesia_football:football123@postgres:5432/ayo_indonesia_football?sslmode=disable" \
	down

## Check Redis
redis-check:
	docker exec -it football-api-redis-1 redis-cli

run-app:
	docker-compose up web worker

