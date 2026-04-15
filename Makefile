# Load .env into Make and into recipe environments (GNU Make).
ifneq (,$(wildcard .env))
-include .env
export
endif

.PHONY: help run build up down db logs clean

run:
	go run .

build:
	mkdir -p bin
	go build -o bin/scheduler .

clean:
	rm -rf bin/

db:
	docker compose up -d db

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f
