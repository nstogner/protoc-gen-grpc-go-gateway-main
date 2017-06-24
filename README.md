# Golang gRPC-HTTP Gateway main.go Generator

Use this protoc plugin to generate a the main.go for running a gRPC gateway server.

```sh
# Install
go get github.com/nstogner/protoc-gen-grpc-go-gateway-main

# This assumes that $GOPATH/bin is a part of $PATH
protoc --grpc-go-gateway-maine_out=./grpcd example.proto
```
