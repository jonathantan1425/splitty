# FROM debian:stretch-slim

# WORKDIR /

# RUN apt-get update && apt-get install -y ca-certificates

# ADD bin /bin/

# CMD ["/bin/sh"]

FROM golang:1.21

WORKDIR /app

COPY * ./

RUN go mod download

RUN --mount=type=cache,target=/go/pkg/mod \
--mount=type=cache,target=/root/.cache/go-build \
go build -o split_my_bills_bot .

EXPOSE 8080

CMD ["./split_my_bills_bot"]
