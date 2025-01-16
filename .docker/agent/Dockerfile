FROM golang:1.23

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o cmd/agent/agent cmd/agent/main.go

CMD ["./cmd/server/agent"]