# Base stage

FROM golang:1.14-alpine as base_block

RUN apk add --no-cache git protobuf

WORKDIR /go/src/github.com/erkrnt/symphony

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go get -d -v ./...
RUN go install github.com/golang/protobuf/protoc-gen-go

COPY ./api ./api

RUN protoc --proto_path="$(pwd)" api/*.proto --go_out=plugins=grpc:../../../

COPY ./cmd/cli ./cmd/cli
COPY ./cmd/block ./cmd/block
COPY ./internal/block ./internal/block
COPY ./internal/cli ./internal/cli
COPY ./internal/service ./internal/service
COPY ./internal/utils ./internal/utils

RUN go install -v ./...

# Final stage

FROM alpine:latest

RUN apk add --no-cache ca-certificates lvm2 open-iscsi

WORKDIR /usr/local/bin

COPY --from=base_block /go/bin/cli .
COPY --from=base_block /go/bin/block .

CMD [ "block" ]
