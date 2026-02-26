# Local testing: start stack, seed event source, run service.

.PHONY: up down run seed test

up:
	docker compose up -d
	@echo "Waiting for MySQL..."
	@sleep 12

down:
	docker compose down

seed:
	docker compose exec -T mysql mysql -u user -ppassword users_metadata -e "\
	INSERT IGNORE INTO event_sources (id, topic_name, enabled, created_at) \
	SELECT UUID(), 'user.events', 1, NOW() FROM DUAL WHERE NOT EXISTS (SELECT 1 FROM event_sources WHERE topic_name = 'user.events'); \
	"
	@echo "Event source 'user.events' ready."

run: up seed
	@echo "Starting service (Ctrl+C to stop)..."
	go run ./cmd

test: up
	@sleep 12
	go build -o /dev/null ./...
	@echo "Build OK. Run 'make run' to start the service."
