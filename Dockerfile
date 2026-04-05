# ==========================================
# STAGE 1: Builder
# ==========================================
# Use the official Go image matching your version
FROM golang:1.25-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules manifests first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go binary. 
# CGO_ENABLED=0 makes it a statically linked binary (no external C dependencies).
# GOOS=linux ensures it compiles for Linux, even if you run this on Windows.
RUN CGO_ENABLED=0 GOOS=linux go build -o gopher-guard ./cmd/server

# ==========================================
# STAGE 2: Production Runner
# ==========================================
# Use a highly minimal, secure Alpine Linux image
FROM alpine:latest

# Add timezone data and CA certificates (standard for web servers)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /app/gopher-guard .

# Create a directory for the Raft BoltDB files
RUN mkdir -p /root/raft-data

# Set the default command to execute the binary
ENTRYPOINT ["./gopher-guard"]