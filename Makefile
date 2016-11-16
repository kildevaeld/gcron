
BUILDDIR?=${PWD}/bin
VERSION?=0.0.1

.PHONY: build clean

build: ${BUILDDIR}/gcron

build-alpine:
	mount=/src/github.com/kildevaeld/gcron
	docker run -v ${PWD}:/go/${mount} -w /go/${mount} -v gobuilder:/go/src kildevaeld/go-builder sh -c "BUILDDIR=bin/alpine make"

update:
	glide update

clean:
	rm -f ${BUILDDIR}/gcron




${BUILDDIR}/gcron: main.go interpreters.go internal/cron.go internal/cronjob.go internal/wrap.go
	
	go build -ldflags "-s -w -X main.VERSION=${VERSION}" -tags notto -o $@ *.go


	
