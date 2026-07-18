.PHONY: up down logs ps config compact-now

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

config:
	docker compose config

compact-now:
	docker compose run --rm spark-compaction spark-submit /opt/spark/work-dir/compact.py
