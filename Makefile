.DEFAULT_GOAL := test

.PHONY: proto
proto: ## Build protos
	./buf.gen.yaml

.PHONY: test
test:
	go test -timeout 10m -v -p=1 -count=1 -race ./...