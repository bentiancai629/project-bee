build:
	go build -o ./bin/project-bee

run: build
	./bin/project-bee

test:
	go test ./...