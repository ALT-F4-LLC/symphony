# Base stage

FROM golang:1.14-alpine

RUN apk add --no-cache git protobuf

WORKDIR /go/src/github.com/erkrnt/symphony

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go get -d -v ./...
RUN go install github.com/golang/protobuf/protoc-gen-go

COPY ./api ./api

RUN protoc --proto_path="$(pwd)" api/*.proto --go_out=plugins=grpc:../../../

COPY ./cmd/cli ./cmd/cli
COPY ./internal/cli ./internal/cli
COPY ./cmd/apiserver ./cmd/apiserver
COPY ./internal/apiserver ./internal/apiserver
COPY ./internal/pkg/cluster ./internal/pkg/cluster
COPY ./internal/pkg/config ./internal/pkg/config
COPY ./internal/pkg/state ./internal/pkg/state

RUN go install -v ./...

# Final stage

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /usr/local/bin

COPY --from=0 /go/bin/cli .
COPY --from=0 /go/bin/apiserver .

CMD [ "apiserver" ]
