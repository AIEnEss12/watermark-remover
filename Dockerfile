# Get Go 1.26 binaries from Bookworm
FROM golang:1.26-bookworm AS go-bin

# Build stage on Trixie (for modern OpenCV 4.7+)
FROM debian:trixie AS builder

# Copy Go from the other stage
COPY --from=go-bin /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"
ENV CGO_ENABLED=1
ENV GOTOOLCHAIN=local

# Install OpenCV dependencies (Trixie)
RUN apt-get update && apt-get install -y \
    build-essential \
    libopencv-dev \
    libwebp-dev \
    libavif-dev \
    pkg-config \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download && go mod tidy

# Copy source code
COPY . .
RUN go mod tidy

# Build the application
RUN go build -o server cmd/server/main.go

# Runtime stage
FROM debian:trixie-slim

# Install runtime dependencies (Trixie)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopencv-photo410 \
    libopencv-imgproc410 \
    libopencv-imgcodecs410 \
    libopencv-videoio410 \
    libopencv-core410 \
    libopencv-highgui410 \
    libopencv-video410 \
    libopencv-features2d410 \
    libopencv-objdetect410 \
    libopencv-dnn410 \
    libopencv-ml410 \
    libopencv-calib3d410 \
    libwebp7 \
    libavif16 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binaries and assets
COPY --from=builder /app/server .
COPY logo.png .

# Environment variables
ENV PORT=8000
ENV GIN_MODE=release

EXPOSE 8000

CMD ["./server"]
