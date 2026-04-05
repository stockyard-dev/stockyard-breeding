build:
	CGO_ENABLED=0 go build -o breeding ./cmd/breeding/

run: build
	./breeding

test:
	go test ./...

clean:
	rm -f breeding

.PHONY: build run test clean
