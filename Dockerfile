FROM rust:1.86-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
        libopencv-dev \
        libclang-dev  \
        clang         \
        pkg-config    \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN cargo build --release --bin server --bin tester

FROM debian:bookworm-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
        libopencv-dev    \
        ca-certificates  \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /app/target/release/server .
COPY --from=builder /app/target/release/tester .
COPY logo.png .

ENV PORT=8000
EXPOSE 8000
CMD ["./server"]
