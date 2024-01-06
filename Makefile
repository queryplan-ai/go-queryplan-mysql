
.PNONY: build
build:
	@echo "Building go-queryplan-mysql..."
	go build \
		-tags "$(BUILDTAGS)" \
		.

.PHONY: test
test: build
	go test -v ./...
