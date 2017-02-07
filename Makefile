
BUILDDIR?=${PWD}/bin
VERSION?=0.0.3
MOUNT=src/github.com/kildevaeld/gcron
.PHONY: build clean update build-alpine

build: ${BUILDDIR}/gcron

build-alpine:
	docker run -v ${PWD}:/go/${MOUNT} -w /go/${MOUNT} -v gobuilder:/go/src kildevaeld/go-builder sh -c "BUILDDIR=bin/alpine make"

update:
	glide update

clean:
	rm -f ${BUILDDIR}/gcron
	rm -f ${BUILDDIR}/gcron/alpine




${BUILDDIR}/gcron: main.go interpreters.go internal/cron.go internal/cronjob.go internal/wrap.go
	go build -ldflags "-s -w -X main.VERSION=${VERSION}" -tags notto -o $@ *.go


	
