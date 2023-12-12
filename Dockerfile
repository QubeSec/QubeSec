# Use the Alpine version of the Golang base image
FROM golang:1.21-alpine

WORKDIR /home/qubesec

# Install build dependencies
RUN apk --no-cache add build-base cmake openssl-dev git

# Get liboqs
RUN git clone --depth 1 --branch main https://github.com/open-quantum-safe/liboqs

# Install liboqs
RUN cmake -S liboqs -B liboqs/build -DBUILD_SHARED_LIBS=ON && \
    cmake --build liboqs/build --parallel 4 && \
    cmake --build liboqs/build --target install

# Configure liboqs-go
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig
ENV LD_LIBRARY_PATH=/usr/local/lib

# Copy the Go Modules manifests
COPY go.mod go.sum ./

# Download Go dependencies
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# Build the Go application
RUN go build -o /manager cmd/main.go

# Set a non-root user for running the application
USER 65532:65532

# Set the entry point for the container
ENTRYPOINT ["/manager"]
