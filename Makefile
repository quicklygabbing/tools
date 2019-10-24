GOPATH ?= $(HOME)/go
LINTER=$(GOPATH)/bin/golangci-lint

.PHONY: deps scripts/githooks lint

$(LINTER):
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1

deps: scripts/githooks
	GO111MODULE=on go mod graph
	GO111MODULE=on go mod vendor

lint: $(LINTER)
	GO111MODULE=on $(GOPATH)/bin/golangci-lint run

test:

.git/hooks:
	mkdir -p .git/hooks

.git/hooks/pre-push: scripts/githooks/pre-push
	cp scripts/githooks/pre-push .git/hooks/pre-push

.git/hooks/pre-commit: scripts/githooks/pre-commit
	cp scripts/githooks/pre-commit .git/hooks/pre-commit

scripts/githooks: .git/hooks .git/hooks/pre-push .git/hooks/pre-commit $(LINTER)
	chmod +x .git/hooks/*
