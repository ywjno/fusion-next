# build frontend
FROM oven/bun:1.3.3-alpine AS fe
WORKDIR /src
COPY .git .git/
COPY frontend ./frontend

# Get version from git
RUN apk add --no-cache git && \
    if [ -n "$FUSION_VERSION" ]; then \
        VERSION="$FUSION_VERSION"; \
    elif git describe --tags --abbrev=0 >/dev/null 2>&1; then \
        VERSION=$(git describe --tags --abbrev=0); \
    else \
        VERSION=$(git rev-parse --short HEAD); \
    fi && \
    echo "Using fusion version string: ${VERSION}" && \
    cd frontend && \
    bun i && \
    VITE_FUSION_VERSION="${VERSION}" bun run build

# build backend
FROM golang:1.25.4-alpine3.22 AS be
# Add Arguments for target OS and architecture (provided by buildx)
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY . ./
COPY --from=fe /src/frontend/build ./frontend/build/

# Build backend
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -ldflags '-extldflags "-static"' \
    -o ./build/fusion \
    ./cmd/server

# deploy
FROM alpine:3.22
WORKDIR /fusion
COPY --from=be /src/build/fusion ./
EXPOSE 8080
RUN mkdir /data
ENV DB="/data/fusion.db"
CMD [ "./fusion" ]
