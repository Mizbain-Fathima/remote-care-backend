FROM golang:1.22

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o api ./cmd/api

EXPOSE 8080 50051

CMD ["./api"]
