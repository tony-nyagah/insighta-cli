FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build static binary named `insighta` (compile the single main file)
RUN CGO_ENABLED=0 GOOS=linux go build -o insighta ./main.go

# --- final image
FROM alpine:3.21
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/insighta .

ENTRYPOINT ["./insighta"]
CMD ["--help"]
