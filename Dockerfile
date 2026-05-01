# Fields2Cover v2 — Go API + C++ gRPC shim, supervised in one container.
#
# Stages:
#   build-cpp  — compiles libFields2Cover.so + f2c-grpc-server
#   build-go   — compiles the Go API binary
#   runtime    — debian-slim base with both binaries, runtime .so deps,
#                openapi.yaml, and a bash supervisor entrypoint
#
# Build:   docker build -t f2c:v2 .
# Run:     docker run -d -p 8080:8080 f2c:v2
# Smoke:   curl -f http://localhost:8080/healthz

# =============================================================================
# Stage 1 — build the C++ library + gRPC shim
# =============================================================================
FROM ubuntu:24.04 AS build-cpp

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        cmake \
        ninja-build \
        git \
        wget \
        pkg-config \
        libboost-dev \
        libgeos-dev \
        libgdal-dev \
        libtbb-dev \
        libeigen3-dev \
        libtinyxml2-dev \
        nlohmann-json3-dev \
        libprotobuf-dev \
        libgrpc++-dev \
        protobuf-compiler \
        protobuf-compiler-grpc \
    && rm -rf /var/lib/apt/lists/*

# NOTE: or-tools is intentionally NOT preinstalled. The top-level
# CMakeLists pulls it in via FetchContent at configure time, picking the
# version we test against. Preinstalling collides with system protobuf
# because the upstream tarball ships its own libcmake configs.

WORKDIR /src
COPY CMakeLists.txt ./
COPY src/ ./src/
COPY include/ ./include/
COPY cmake/ ./cmake/
COPY swig/ ./swig/
COPY proto/ ./proto/
COPY f2c-grpc/ ./f2c-grpc/
COPY LICENSE ./
COPY README.rst ./
COPY package.xml ./

RUN cmake -S . -B build \
        -GNinja \
        -DBUILD_PYTHON=OFF \
        -DBUILD_TUTORIALS=OFF \
        -DBUILD_TESTING=OFF \
        -DBUILD_DOC=OFF \
        -DBUILD_F2C_GRPC=ON \
        -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build --target Fields2Cover f2c-grpc-server -j

# Copy the artifacts to a stable location for the runtime stage to grab.
# Both libFields2Cover.so AND every .so produced by the FetchContent
# vendored deps (or-tools, steering_functions, matplot, nodesoup, ...)
# need to be on the runtime image's ldconfig path. Sweep the build tree.
RUN mkdir -p /artifacts/lib /artifacts/bin \
    && cp -P build/libFields2Cover.so* /artifacts/lib/ \
    && cp build/f2c-grpc/f2c-grpc-server /artifacts/bin/ \
    && find build/_deps -name '*.so' -o -name '*.so.*' \
        | xargs -I{} cp -P {} /artifacts/lib/

# =============================================================================
# Stage 2 — build the React SPA
# =============================================================================
FROM node:22-alpine AS build-frontend

WORKDIR /ui

# Copy manifests first so the npm install layer is cached independently
# of source changes (invalidated only when package-lock.json changes).
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline

COPY frontend/ ./
RUN npm run build
# Output: /ui/dist/

# =============================================================================
# Stage 3 — build the Go API binary
# =============================================================================
FROM golang:1.22-bookworm AS build-go

WORKDIR /src
COPY api-go/go.mod api-go/go.sum ./
RUN go mod download

COPY api-go/ ./
# Overwrite the committed placeholder dist/ with the real React build.
# Must come AFTER "COPY api-go/ ./" (which restores the placeholder from
# git) and BEFORE "go build" (which bakes dist/ into the binary via go:embed).
COPY --from=build-frontend /ui/dist/ ./internal/static/dist/
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/api ./cmd/api

# =============================================================================
# Stage 4 — runtime
# =============================================================================
FROM ubuntu:24.04 AS runtime

ENV DEBIAN_FRONTEND=noninteractive

# Runtime shared-library deps for the C++ shim. Mirror libgdal34t64,
# libgeos, libtbb, libtinyxml2, libprotobuf, libgrpc++ — anything the
# build-cpp stage linked against. No -dev packages, no compilers.
RUN apt-get update && apt-get install -y --no-install-recommends \
        ca-certificates \
        libgdal34t64 \
        libgeos-c1t64 \
        libtbb12 \
        libtinyxml2-10 \
        libprotobuf32t64 \
        libgrpc++1.51t64 \
        libgomp1 \
        bash \
        curl \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 10001 --create-home --shell /usr/sbin/nologin f2c

# C++ artifacts
COPY --from=build-cpp /artifacts/lib/ /usr/local/lib/
COPY --from=build-cpp /artifacts/bin/f2c-grpc-server /usr/local/bin/f2c-grpc-server
RUN ldconfig

# Go binary + spec for /docs
COPY --from=build-go /out/api /usr/local/bin/api
COPY api-go/openapi.yaml /etc/f2c/openapi.yaml

# Supervisor entrypoint
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER f2c
ENV F2C_GRPC_SOCKET=/tmp/f2c.sock \
    API_PORT=8080 \
    OPENAPI_PATH=/etc/f2c/openapi.yaml

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
