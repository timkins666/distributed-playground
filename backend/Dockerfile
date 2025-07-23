ARG GO_VERSION='x'
FROM golang:${GO_VERSION}-alpine

ARG SERVICE_NAME

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./pkg/. ./pkg/
COPY ./svc/${SERVICE_NAME}/. ./svc/${SERVICE_NAME}/

WORKDIR /app/svc/${SERVICE_NAME}

RUN go build -o service

CMD ["./service"]