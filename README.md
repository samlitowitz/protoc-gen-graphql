# protoc-gen-graphql
[![Go Report Card](https://goreportcard.com/badge/github.com/samlitowitz/protoc-gen-graphql)](https://goreportcard.com/report/github.com/samlitowitz/protoc-gen-graphql)
[![GoDoc](https://godoc.org/github.com/samlitowitz/protoc-gen-graphql/pkg/protoc-gen-graphql/generator?status.svg)](https://godoc.org/github.com/samlitowitz/protoc-gen-graphql/pkg/protoc-gen-graphql/generator)

# Installation
Follow the steps outlined [here](https://github.com/golang/protobuf).
To install `protoc-gen-graphql` run `go get -u github.com/samlitowitz/protoc-gen-graphql/cmd/protoc-gen-graphql`. 

# Usage with protoc
`protoc --graphql_out=. *.proto`

# Supported

Supports enumerations, nested definitions, and the repeated keyword. 
Types not defined in the file are converted to scalar, e.g. `google.protobuf.Timestamp` maps to `GoogleProtobufTimestamp`.

## Scalar Values Mapping
| GraphQL | Protobuf (v3) | 
| --- | --- |
| Int | int32, sfixed32, sint32 |
| Float | float |
| String | bytes, string |
| Boolean | bool |
| ID | string (but not really, this is unmapped) |
| scalar Float64 | double |
| scalar Int64 | int64, sfixed64, sint64 |
| scalar Uint32 | fixed32, uint32 |
| scalar Uint64 | fixed64, uint64 | 
