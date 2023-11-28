FROM golang:1.21

WORKDIR /home/oqs

# Install dependencies
RUN apt update && \
    apt install -y build-essential cmake libssl-dev

# Get liboqs
RUN git clone --depth 1 --branch main https://github.com/open-quantum-safe/liboqs

# Install liboqs
RUN cmake -S liboqs -B liboqs/build -DBUILD_SHARED_LIBS=ON && \
    cmake --build liboqs/build --parallel 4 && \
    cmake --build liboqs/build --target install

# Get liboqs-go
RUN git clone --depth 1 --branch main https://github.com/open-quantum-safe/liboqs-go

# Configure liboqs-go
ENV PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/home/oqs/liboqs-go/.config
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

RUN go build -a -o /manager cmd/main.go

USER 65532:65532

ENTRYPOINT ["/manager"]
