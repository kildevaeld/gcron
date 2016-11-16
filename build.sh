#!/bin/sh

THIS="${PWD/$GOPATH/}"
MOUNT="/go${THIS}"
docker run -ti --rm  -v ${PWD}:${MOUNT} -w ${MOUNT} kildevaeld/go-builder sh -c make