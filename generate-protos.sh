#!/bin/bash

rm -rf protos/
protoc -I=. --go_out=plugins=grpc:. protos.proto
cp -r github.com/ducc/kwɒnt/protos protos
rm -rf github.com

python3 -m grpc_tools.protoc -I. --python_out=strategy_evaluator/. --grpc_python_out=strategy_evaluator/. protos.proto
