FROM golang:1.23

WORKDIR /app
COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@latest; \
    go mod tidy 
RUN go build -o cmd/server/server cmd/server/main.go 

CMD ["./cmd/server/server"]