FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build a statically linked binary with CGO disabled
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o analytics-server .

# Use scratch as the base image for the smallest possible footprint
FROM scratch

WORKDIR /app

COPY --from=builder /app/analytics-server .

# The analytics server will be accessible on port 8080
EXPOSE 8080

# You can override this with -e ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
ENV ALLOWED_ORIGINS="*"

CMD ["./analytics-server"] 