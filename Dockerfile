FROM golang:1.15 as builder

EXPOSE 8082
EXPOSE 8080

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go build -v -o todo . 

FROM debian:buster-slim

COPY --from=builder /go/src/app/todo .

RUN ls -la ./

CMD ["./todo", "grpc"]

