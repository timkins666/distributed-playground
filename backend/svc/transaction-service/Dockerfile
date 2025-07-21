ARG GO_VERSION='x'
FROM golang:${GO_VERSION}-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/svc/transaction-service

# RUN CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
# RUN CGO_ENABLED=0 go build -gcflags "all=-N -l" -o go-debugger .

RUN go build -o transaction-service

# delve port
EXPOSE 4000 

CMD ["./transaction-service"]

# CMD [ "/go/bin/dlv", "--listen=:4000", "--headless=true", "--log=true", "--accept-multiclient", "--api-version=2", "exec", "./auth-service" ]

