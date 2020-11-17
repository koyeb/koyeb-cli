#!/bin/bash

curl -s https://developer.koyeb.com/public.swagger.json > ./public.swagger.json
curl -s https://developer.koyeb.com/stackv1.swagger.build.json > ./stackv1.swagger.build.json
CLIENT_OUT=./pkg/gen/kclient
rm -rf ${CLIENT_OUT}
mkdir -p ${CLIENT_OUT}

## This script pulls from the api doc the latest swagger and creates a client from it
docker run --rm -v `pwd`:/go/src/github.com/koyeb/koyeb-cli -w /go/src/github.com/koyeb/koyeb-cli quay.io/goswagger/swagger:v0.23.0 generate client -f public.swagger.json -t ${CLIENT_OUT}
