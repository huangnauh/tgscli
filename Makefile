APP=tgscli
REPO_PATH=github.com/huangnauh/${APP}
VERSION_IMPORT=${REPO_PATH}/pkg/version
GIT_COMMIT=$(shell git rev-parse --short HEAD)
GIT_DESCRIBE=$(shell git describe --tags --always)
GOLDFLAGS=-X $(VERSION_IMPORT).GitCommit=$(GIT_COMMIT) -X $(VERSION_IMPORT).GitDescribe=$(GIT_DESCRIBE)

build:
	go build -gcflags=all="-N -l" -ldflags '$(GOLDFLAGS)' ./cmd/${APP}

docs:
	./tgscli docs

install:
	go install -ldflags '$(GOLDFLAGS)' ./cmd/${APP}

release-skip:
	VERSION_IMPORT=${VERSION_IMPORT} GIT_COMMIT=${GIT_COMMIT} GIT_DESCRIBE=${GIT_DESCRIBE} \
		goreleaser --snapshot --skip-publish --rm-dist

release:
	VERSION_IMPORT=${VERSION_IMPORT} GIT_COMMIT=${GIT_COMMIT} GIT_DESCRIBE=${GIT_DESCRIBE} \
		goreleaser --rm-dist

.PHONY: build docs install release-build
