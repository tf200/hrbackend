## OR-Tools Go bindings require CGO + native OR-Tools libraries.
## We build on Debian (glibc) and vendor OR-Tools binary distribution.

# Stage 1: Builder
FROM golang:1.26.1-bookworm AS builder

WORKDIR /app

ARG ORTOOLS_TAG=v9.12
ARG ORTOOLS_VERSION=9.12.4544
ARG ORTOOLS_DIST=debian-12

# System deps for CGO + downloading OR-Tools + direct git module fetches.
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        git \
        wget \
        build-essential \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Copy the go.mod and go.sum files first to cache dependencies
COPY go.mod go.sum ./

# The main module uses a local replace for OR-Tools Go bindings.
COPY third-party/google-or-tools/go.mod ./third-party/google-or-tools/go.mod


ENV GOPROXY=https://proxy.golang.org,direct

# OR-Tools module is very large; proxy fetch can fail.
ENV GONOPROXY=github.com/google/or-tools
ENV GONOSUMDB=github.com/google/or-tools

# Download dependencies
RUN go mod download -x

# Install OR-Tools (native headers + shared libs)
RUN wget -O /tmp/ortools.tar.gz -q "https://github.com/google/or-tools/releases/download/${ORTOOLS_TAG}/or-tools_amd64_${ORTOOLS_DIST}_cpp_v${ORTOOLS_VERSION}.tar.gz" \
    && mkdir -p /opt/ortools \
    && tar -xzf /tmp/ortools.tar.gz -C /opt/ortools --strip-components=1 \
    && rm -f /tmp/ortools.tar.gz

# Copy the rest of the application source code
COPY . .

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-I/opt/ortools/include"
ENV CGO_LDFLAGS="-L/opt/ortools/lib -lortools -Wl,-rpath,/opt/ortools/lib"

# Build the Go application
RUN go build -tags ortools -ldflags="-s -w" -o main .

# Stage 2: Final Image
FROM debian:bookworm-slim


# Install runtime deps (certs, tzdata, C++ runtime libs)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata \
        libstdc++6 \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Copy OR-Tools shared libraries
COPY --from=builder /opt/ortools/lib /opt/ortools/lib

ENV LD_LIBRARY_PATH=/opt/ortools/lib

COPY migrations /app/migrations

# Expose the desired port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/main"]
