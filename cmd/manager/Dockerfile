FROM golang:1.14-alpine as base_manager

RUN apk add --no-cache git protobuf

WORKDIR /go/src/github.com/erkrnt/symphony

COPY ./go.mod ./go.mod

COPY ./go.sum ./go.sum

RUN go get -d -v ./...

RUN go install github.com/golang/protobuf/protoc-gen-go

COPY ./api ./api

RUN protoc --proto_path="$(pwd)" api/*.proto --go_out=plugins=grpc:../../../

COPY ./cmd/cli ./cmd/cli

COPY ./cmd/manager ./cmd/manager

COPY ./internal/cli ./internal/cli

COPY ./internal/manager ./internal/manager

COPY ./internal/pkg/config ./internal/pkg/config

COPY ./internal/pkg/gossip ./internal/pkg/gossip

COPY ./internal/pkg/state ./internal/pkg/state

RUN go install -v ./...

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /usr/local/bin

COPY --from=base_manager /go/bin/cli .

COPY --from=base_manager /go/bin/manager .

CMD [ "manager" ]
