FROM golang:1.23

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build cmd/server/main.go 

CMD ["./cmd/server/server"]