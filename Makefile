.PHONY: build run clean docker-up docker-down fmt test

build:
	go build -o rsshub ./cmd/

run: build
	./rsshub

clean:
	rm -f rsshub

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down -v

fmt:
	gofumpt -l -w .

test:
	go test -v -race ./...

install-tools:
	go install mvdan.cc/gofumpt@latest