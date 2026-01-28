# Build arguments
ARG DEBUG=false
ARG COVER=false
ARG BUILDPLATFORM

# ------------------
# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder
ARG DEBUG
ARG COVER
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache git

# Install conditional debug dependencies
RUN if [ "$DEBUG" = "true" ]; then \
  go install github.com/go-delve/delve/cmd/dlv@latest; \
else \
  mkdir -p /go/bin; \
  touch /go/bin/dlv; \
fi

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Conditionally build with debug or coverage flags
RUN if [ "$DEBUG" = "true" ]; then \
  echo "Building DEBUG binary"; \
  CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -gcflags="all=-N -l" -o jam-build-propsdb ./cmd/server; \
elif [ "$COVER" = "true" ]; then \
  echo "Building COVER binary"; \
  CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -cover -coverpkg=./... -covermode=atomic -o jam-build-propsdb ./cmd/server; \
else \
  echo "Building PRODUCTION binary"; \
  CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -o jam-build-propsdb ./cmd/server; \
fi

# Build the healthcheck application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -o healthcheck ./cmd/healthcheck

# ------------------
# Runtime stage
FROM alpine:3.21 AS runtime

# Install ca-certificates for HTTPS and wget for health checks
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binaries from builder
COPY --from=builder --chown=appuser:appuser /app/jam-build-propsdb .
COPY --from=builder --chown=appuser:appuser /app/healthcheck .
# Copy dlv if it was built
COPY --from=builder /go/bin/dlv* /usr/local/bin/

# Create coverage directory
RUN mkdir -p /app/coverage && \
  chown -R appuser:appuser /app/coverage

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check using the healthcheck binary
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/app/healthcheck"]

# Run the application
CMD ["./jam-build-propsdb"]

