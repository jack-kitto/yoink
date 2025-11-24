APP_NAME=yoink

build:
	go build -o bin/$(APP_NAME) main.go

run:
	go run main.go

fmt:
	go fmt ./...

clean:
	rm -rf bin

install:
	go install ./...

test:
	go test ./...
