APP=tgscli
REPO_PATH=github.com/huangnauh/${APP}
VERSION_IMPORT=${REPO_PATH}/version
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_DESCRIBE=$(shell git describe --tags --always)
GOLDFLAGS=-X $(VERSION_IMPORT).GitCommit=$(GIT_COMMIT) -X $(VERSION_IMPORT).GitDescribe=$(GIT_DESCRIBE)

default: tidy build

build:
	go build -gcflags=all="-N -l" -ldflags '$(GOLDFLAGS)' ./cmd/${APP}

tidy:
	go mod tidy

docs:
	tgscli docs

install-app:
	go install -ldflags '$(GOLDFLAGS)' ./cmd/${APP}

install: tidy install-app docs

release-skip:
	VERSION_IMPORT=${VERSION_IMPORT} GIT_COMMIT=${GIT_COMMIT} GIT_DESCRIBE=${GIT_DESCRIBE} \
		goreleaser --snapshot --skip-publish --rm-dist

release:
	VERSION_IMPORT=${VERSION_IMPORT} GIT_COMMIT=${GIT_COMMIT} GIT_DESCRIBE=${GIT_DESCRIBE} \
		goreleaser --rm-dist

.PHONY: build tidy docs install release-build
