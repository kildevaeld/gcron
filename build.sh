#!/bin/sh

THIS="${PWD/$GOPATH/}"

echo ${THIS}
docker run -ti --rm  -v ${PWD}:/go${THIS} kildevaeld/go-builder sh 