.PHONY: db-up mc
db-up:
	docker compose up -d

mc:
	migrate create -ext sql -dir ./migrations -seq ${NAME}
