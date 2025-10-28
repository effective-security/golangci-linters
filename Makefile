include .project/gomod-project.mk
BUILD_FLAGS=

.PHONY: *

.SILENT:

default: help

all: clean tools generate change_log covtest

#
# clean produced files
#
clean:
	go clean ./...
	rm -rf \
		${COVPATH} \
		${PROJ_BIN}

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
	go install github.com/effective-security/cov-report/cmd/cov-report@latest

change_log:
	echo "Recent changes" > ./change_log.txt
	echo "Build Version: $(GIT_VERSION)" >> ./change_log.txt
	echo "Commit: $(GIT_HASH)" >> ./change_log.txt
	echo "==================================" >> ./change_log.txt
	git log -n 20 --pretty=oneline --abbrev-commit >> ./change_log.txt

build:
	echo "*** Building linters"
	CGO_ENABLED=1 go build -buildmode=plugin -o ${PROJ_ROOT}/bin/custom-linters-v2.5.0.so ./cmd/custom-linters
