.PHONY: start stop restart rebuild logs clean

start:
	docker-compose up -d

stop:
	docker-compose down

restart:
	docker-compose restart

rebuild:
	docker-compose build --no-cache
	docker-compose up -d

logs:
	docker-compose logs -f matching-engine

clean:
	docker-compose down -v
	docker system prune -f