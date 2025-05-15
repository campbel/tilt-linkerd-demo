.PHONY: proto
proto:
	cd proto && buf generate

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test ./...

.PHONY: run
run:
	tilt up