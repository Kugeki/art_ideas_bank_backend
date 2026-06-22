.PHONY: run, run-and-attach, stop, remove-dangling, install-goose, generate-migration, install-swag, generate-swagger

KNOWN_TARGETS = target
ARGS := $(filter-out $(KNOWN_TARGETS),$(MAKECMDGOALS))

run:
	docker compose up -d --build

run-and-attach:
	docker compose up --build

stop:
	docker compose down

remove-dangling:
	docker rmi $(docker images --filter "dangling=true" -q --no-trunc)

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest

generate-migration:
	chmod +x ./migrations/generate-migration.sh
	(cd migrations ; ./generate-migration.sh $(name))

install-swag:
	go install github.com/swaggo/swag/cmd/swag@latest

generate-swagger:
	swag fmt -d internal/ports/restapi
	swag init -d internal/ports/restapi -g restapi.go
