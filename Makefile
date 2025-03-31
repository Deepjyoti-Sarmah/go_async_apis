db_login:
	psql ${DATABASE_URL}

db_create_migration:
	migrate create -ext sql -dir migrations -seq ${name}

db_fix_migration:
	migrate -database ${DATABASE_URL} -path migrations force 1

db_migrate:
	migrate -database ${DATABASE_URL} -path migrations up
