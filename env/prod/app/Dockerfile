FROM golang:1.23-bookworm AS builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN apt-get update && apt-get install -y unzip
ENV PATH="/root/.bun/bin:${PATH}"
RUN LDFLAGS="-ldflags '-X main.Environment=prod'" make build

# Use the official Debian slim image for a lean production container.
# https://hub.docker.com/_/debian
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/bin /app/bin
COPY --from=builder /app/cmd/app/static /app/cmd/app/static

# Run the app on container startup.
CMD ["/app/bin/app"]