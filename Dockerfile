FROM golang:latest

WORKDIR /app

COPY go.mod ./
RUN go mod download

CMD CGO_ENABLED=0 go build -o mysqlbr

