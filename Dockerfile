FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o analytics-server .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/analytics-server .

EXPOSE 8080

CMD ["./analytics-server"] 