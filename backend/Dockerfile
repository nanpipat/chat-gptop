FROM golang:1.24

WORKDIR /app

COPY backend/ .

RUN go mod tidy
RUN go build -o server cmd/server/main.go

CMD ["./server"]
