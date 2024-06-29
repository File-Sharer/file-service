dev:
	go run cmd/app/main.go

build:
	go build -o cmd/app/app cmd/app/main.go

run:
	./cmd/app/app
