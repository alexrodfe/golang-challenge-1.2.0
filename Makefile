tidy ::
	@go mod tidy && go mod vendor

mocks ::
	@go run github.com/vektra/mockery/v3@v3.8.0

seed ::
	@go run cmd/seed/main.go

run ::
	@go run cmd/server/main.go

test ::
	@go test -v -count=1 -race ./... -coverprofile=coverage.out -covermode=atomic

docker-up ::
	docker compose up -d

docker-down ::
	docker compose down
