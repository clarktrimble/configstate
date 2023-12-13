EXECS   := $(wildcard cmd/*)
TARGETS := ${EXECS:cmd/%=%}

TESTA   := ${shell go list ./... | grep -v /cmd/ | grep -v /test/ | grep -v /mock}

BRANCH   := ${shell git branch --show-current}
REVCNT   := ${shell git rev-list --count $(BRANCH)}
REVHASH  := ${shell git log -1 --format="%h"}

LDFLAGS  := -X main.version=${BRANCH}.${REVCNT}.${REVHASH}

all: check clean build

check: gen lint test

cover:
	go test -coverprofile=cover.out ${TESTA} && \
	go tool cover -func=cover.out

gen:
	go generate ./...

lint:
	golangci-lint run ./...

test:
	go test -count 1 ${TESTA}

race:
	CGO_ENABLED=1 go test -count 1 -race ${TESTA}

clean:
	-rm -rf bin

build: ${TARGETS}
	@echo ":: Done"

${TARGETS}:
	@echo ":: Building $@"
	go build -ldflags '${LDFLAGS}' -o bin/$@ cmd/$@/main.go

.PHONY: test discover

