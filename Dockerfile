FROM golang:1.22 AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api ./cmd/api

FROM debian:bullseye
WORKDIR /app
COPY --from=build /app/api /app/api

EXPOSE 8080 50051
CMD ["/app/api"]
