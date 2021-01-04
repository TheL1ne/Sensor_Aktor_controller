# Sensor Aktor Controller
This repository conatins my mockup for my bachelor thesis about the usage of ontologies for defining the normal state of a network with clients and servers and finding anomalies in it. My small example is made up of a sensor, an actor and a controller which do not really sense or act on anything. They are just there to probide network traffic I can test the idea on.

## Dependencies
* [protocol buffers](https://github.com/protocolbuffers/protobuf)
* make

## Starting the mockup

## generation from proto
I use protobuffers to define my small services and take advantage of the easily included GRPC functionality. To regenerate them you need to:
1. install a version of the protobuf compiler from [here](https://github.com/protocolbuffers/protobuf).
2. call `protoc --go_out=plugins=grpc:. api.proto` in the api directory or just do `make proto` if you have make installed