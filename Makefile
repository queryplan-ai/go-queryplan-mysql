
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

	docker rm -f qpperformance || true
	docker run --rm -d \
		--name qpperformance \
		-p 33306:3306 \
		-e MYSQL_ROOT_PASSWORD=qpperformance \
		-e MYSQL_DATABASE=qpperformance \
		-e MYSQL_USER=qpperformance \
		-e MYSQL_PASSWORD=qpperformance \
		mysql:8.2
	@sleep 20
	QUERYPLAN_DB_URI="qpperformance:qpperformance@tcp(localhost:33306)/qpperformance" ./bin/performance

