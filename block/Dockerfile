FROM golang:1.13-alpine as packages
WORKDIR /packages
COPY go.mod .
COPY go.sum .
RUN go mod download

FROM packages as build
COPY . .
RUN go build

FROM alpine
COPY --from=build /packages/block /srv/block
CMD [ "/srv/block" ]