.PHONY: run start stop logs

run:
	docker compose up -d --build --wait
	@echo "Grafana: http://localhost:3000 (admin/admin)"

start: run

stop:
	docker compose down

logs:
	docker compose logs -f reproducer