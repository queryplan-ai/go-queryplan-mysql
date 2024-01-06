
.PNONY: build
build:
	@echo "Building go-queryplan-mysql..."
	go build \
		-tags "$(BUILDTAGS)" \
		.

.PHONY: test
test: build
	go test -v ./...

.PHONY: performance
performance: build
	mkdir -p ./bin
	go build -o ./bin/performance \
		-tags "$(BUILDTAGS)" \
		./performance

	./bin/performance

