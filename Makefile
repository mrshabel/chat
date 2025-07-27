
.PHONY: start-db
start-db: # start the database instance
	docker compose up db -d
