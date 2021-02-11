from golang:alpine  
LABEL maintainer="the_l1ne@gmx.de"
WORKDIR /usr/local/go/src/github.com/TheL1ne/Sensor_Aktor_controller/

RUN apk add build-base

COPY  . /usr/local/go/src/github.com/TheL1ne/Sensor_Aktor_controller/   

RUN go build ./...