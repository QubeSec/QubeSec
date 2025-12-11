# Global build arguments
ARG BASE_ALPINE_VERSION=3.23
ARG BASE_GOLANG_VERSION=1.25
ARG LIBOQS_VERSION=0.15.0
ARG LIBOQS_GO_VERSION=0.12.0
ARG OQS_PROVIDER_VERSION=0.10.0
ARG OPENSSL_VERSION=3.6.0

# Stage 1: Build OpenSSL
FROM alpine:${BASE_ALPINE_VERSION} AS buildopenssl

# Re-declare the global argument for this stage
ARG OPENSSL_VERSION

# Install build dependencies for OpenSSL
RUN apk add build-base linux-headers libtool automake autoconf make git wget

# Clone OpenSSL repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch openssl-${OPENSSL_VERSION} https://github.com/openssl/openssl.git

# Build and install OpenSSL
WORKDIR /optbuild/openssl
RUN LDFLAGS="-Wl,-rpath -Wl,/opt/openssl/lib64" ./config shared --prefix=/opt/openssl && \
    make -j 8 && \
    make install && \
    ln -s /opt/openssl/lib64 /opt/openssl/lib || true && \
    ln -s /opt/openssl/lib /opt/openssl/lib64 || true

# Stage 2: Build liboqs
ARG BASE_ALPINE_VERSION
FROM alpine:${BASE_ALPINE_VERSION} AS buildliboqs

# Re-declare the global argument for this stage
ARG LIBOQS_VERSION

# Install build dependencies for liboqs
RUN apk add build-base linux-headers libtool automake autoconf cmake ninja make git wget

# Clone liboqs repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch ${LIBOQS_VERSION} https://github.com/open-quantum-safe/liboqs.git

# Get OpenSSL image (from cache)
COPY --from=buildopenssl /opt/openssl /opt/openssl

# Build and install liboqs
WORKDIR /optbuild/liboqs
RUN mkdir build && \
    cd build && \
    cmake -G"Ninja" .. -DOPENSSL_ROOT_DIR=/opt/openssl -DOQS_DIST_BUILD=ON -DCMAKE_INSTALL_PREFIX=/opt/liboqs && \
    ninja install

# Stage 3: Build oqs-provider
ARG BASE_ALPINE_VERSION
FROM alpine:${BASE_ALPINE_VERSION} AS buildoqsprovider

# Re-declare the global argument for this stage
ARG OQS_PROVIDER_VERSION

# Install build dependencies for oqs-provider
RUN apk add build-base linux-headers libtool cmake ninja git wget

# Clone oqs-provider repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch ${OQS_PROVIDER_VERSION} https://github.com/open-quantum-safe/oqs-provider.git

# Get openssl and liboqs
COPY --from=buildopenssl /opt/openssl /opt/openssl
COPY --from=buildliboqs /opt/liboqs /opt/liboqs

# Build and install oqs-provider
WORKDIR /optbuild/oqs-provider
RUN liboqs_DIR=/opt/liboqs cmake -DOPENSSL_ROOT_DIR=/opt/openssl -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=/opt/openssl -S . -B _build && \
    cmake --build _build && \
    cmake --install _build && \
    cp _build/lib/oqsprovider.so /opt/openssl/lib64/ossl-modules && \
    sed -i "s/default = default_sect/default = default_sect\noqsprovider = oqsprovider_sect/g" /opt/openssl/ssl/openssl.cnf && \
    sed -i "s/\[default_sect\]/\[default_sect\]\nactivate = 1\n\[oqsprovider_sect\]\nactivate = 1\n/g" /opt/openssl/ssl/openssl.cnf && \
    sed -i "s/providers = provider_sect/providers = provider_sect\nssl_conf = ssl_sect\n\n\[ssl_sect\]\nsystem_default = system_default_sect\n\n\[system_default_sect\]\nGroups = \$ENV\:\:DEFAULT_GROUPS\n/g" /opt/openssl/ssl/openssl.cnf && \
    sed -i "s/HOME\t\t\t= ./HOME           = .\nDEFAULT_GROUPS = kyber768/g" /opt/openssl/ssl/openssl.cnf

# Stage 4: Build operator
ARG BASE_GOLANG_VERSION
FROM golang:${BASE_GOLANG_VERSION}-alpine AS buildoperator

# Re-declare the global argument for this stage
ARG LIBOQS_VERSION
ARG LIBOQS_GO_VERSION

# Set working directory
WORKDIR /home/qubesec

# Install build dependencies
RUN apk add build-base cmake openssl-dev git

# Clone liboqs repository
RUN git clone --depth 1 --branch ${LIBOQS_VERSION} https://github.com/open-quantum-safe/liboqs.git

# Install liboqs
RUN cmake -S liboqs -B liboqs/build -DBUILD_SHARED_LIBS=ON && \
    cmake --build liboqs/build --parallel 4 && \
    cmake --build liboqs/build --target install

RUN git clone --depth=1 --branch ${LIBOQS_GO_VERSION} https://github.com/open-quantum-safe/liboqs-go.git

# Configure liboqs-go
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/home/qubesec/liboqs-go/.config
ENV LD_LIBRARY_PATH=/usr/local/lib

# Copy Go source code
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# Build the Go application
RUN go build -o /manager cmd/main.go

# Stage 5: Final image
ARG BASE_ALPINE_VERSION
FROM alpine:${BASE_ALPINE_VERSION}

# Copy oqs-provider and operator artifacts
COPY --from=buildoqsprovider /opt/openssl /opt/openssl
COPY --from=buildoperator /manager /manager
COPY --from=buildoperator /usr/local/lib /usr/local/lib

# Set environment variables
ENV PATH="/opt/openssl/bin:${PATH}"

# Set user
USER 65532:65532

# Set entry point
ENTRYPOINT ["/manager"]
