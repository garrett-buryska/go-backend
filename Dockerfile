# Step 1: Build the binary with CGO support
FROM golang:1.26-alpine AS builder

# Install gcc and musl-dev so CGO can compile the SQLite driver
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build with CGO_ENABLED=1
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Step 2: Final Image
FROM alpine:latest
# We need libc compatibility for the CGO-linked binary to run
RUN apk --no-cache add ca-certificates libc6-compat
WORKDIR /root/
COPY --from=builder /app/main .
RUN mkdir ./data

EXPOSE 8080
CMD ["./main"]