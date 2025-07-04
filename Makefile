.PHONY: build install clean test

# Build the extension
build:
	go build -o gh-standup ./cmd/standup

# Install the extension locally for development
install: build
	cp gh-standup $(shell gh extension list | grep standup | awk '{print $$3}' || echo ~/.local/share/gh/extensions/gh-standup)/gh-standup

# Clean build artifacts
clean:
	rm -f gh-standup

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Run the extension locally
run:
	go run ./cmd/standup $(ARGS)
