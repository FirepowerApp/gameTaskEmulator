.PHONY: build clean test

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/gameTaskEmulator ./cmd/gameTaskEmulator
	@echo "Built bin/gameTaskEmulator"

clean:
	rm -rf bin

test:
	go test ./...
