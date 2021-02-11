This repository conatins my mockup for my bachelor thesis about the usage of ontologies for defining the normal state of a network with clients and servers and finding anomalies in it. My small example is made up of a sensor, an actor and a controller which do not really sense or act on anything. They are just there to provide network traffic I can test the idea on.
I added the capability of a local sqllite3 database to record my traffic as it is sent and evaluate late on if I foudn all my injected errors.

## Dependencies
* [protocol buffers](https://github.com/protocolbuffers/protobuf)
* make
* everything in `go.mod`

## Starting the mockup

## generation from proto
I use protobuffers to define my small services and take advantage of the easily included GRPC functionality. To regenerate them you need to:
1. install a version of the protobuf compiler from [here](https://github.com/protocolbuffers/protobuf).
2. call `protoc --go_out=plugins=grpc:. api.proto` in the api directory or just do `make proto` if you have make installed