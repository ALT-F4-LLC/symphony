FROM golang:1.14-alpine as base_block

RUN apk add --no-cache git lvm2 protobuf

WORKDIR /go/src/github.com/erkrnt/symphony

COPY ./go.mod ./go.mod

COPY ./go.sum ./go.sum

RUN go get -d -v ./...

RUN go install github.com/golang/protobuf/protoc-gen-go

COPY ./api ./api

RUN protoc --proto_path="$(pwd)" api/*.proto --go_out=plugins=grpc:../../../

COPY ./cmd/block ./cmd/block

COPY ./cmd/cli ./cmd/cli

COPY ./internal/block ./internal/block

COPY ./internal/cli ./internal/cli

COPY ./internal/pkg/config ./internal/pkg/config

COPY ./internal/pkg/gossip ./internal/pkg/gossip

RUN go install -v ./...

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /usr/local/bin

COPY --from=base_block /go/bin/block .

COPY --from=base_block /go/bin/cli .

CMD [ "block" ]