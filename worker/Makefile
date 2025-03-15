.PHONY: up down shell

up:
	docker compose up --build -d

down:
	docker compose down --remove-orphans

shell:
	docker compose exec -it go_server bash

logs:
	docker compose logs -f

restart: down up