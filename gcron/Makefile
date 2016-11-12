
BUILDDIR?=${PWD}/bin
VERSION?=0.0.1
.PHONY: build clean

${BUILDDIR}/gcron: main.go interpreters.go
	
	go build -ldflags "-s -w -X main.VERSION=${VERSION}" -tags notto -o $@ $^

build: ${BUILDDIR}/gcron
	
clean:
	rm -f ${BUILDDIR}/gcron