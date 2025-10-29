FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build \
      -trimpath \
      -ldflags '-s -w -extldflags "-static"' \
      -o /out/app ./cmd/sequentialthinking/main.go

FROM scratch
# Binary
COPY --from=build /out/app /app
# Run as non-root
USER 65532:65532
ENTRYPOINT ["/app"]
