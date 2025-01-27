CREATE DATABASE social;



//initialize migrate
migrate create -seq -ext sql -dir ./cmd/migrate/migrations create_users

//doing migration up
migrate -path=./cmd/migrate/migrations -database="postgres://admin:adminpassword@localhost:5433/socialnetwork?sslmode=disable" up