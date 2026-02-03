# Starts the migration of all un-applied migration files
.PHONY: migrate-up
migrate-up:
	migrate -path ${MIGRATIONS_PATH} -database ${DATABASE_URL} -verbose up


# Pulls down every migration, i.e a full db rollback
.PHONY: migrate-down
migrate-down:
	migrate -path ${MIGRATIONS_PATH} -database ${DATABASE_URL} -verbose down


# Generates a new migration file for the given name
.PHONY: migrate-generate
migrate-generate:
	migrate create -ext sql -dir ${MIGRATIONS_PATH} -seq $(name)

# Generate SQLC models and bindings
.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate
