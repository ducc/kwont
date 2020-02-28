#!/bin/bash

rm -rf protos/
protoc -I=. --go_out=plugins=grpc:. protos.proto 
cp -r github.com/ducc/kwɒnt/protos protos
rm -rf github.com

