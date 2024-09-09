# Stage 1: Build OpenSSL
FROM alpine:3.20 as buildopenssl

# Global build arguments
ARG INSTALLDIR_OPENSSL=/opt/openssl32
ARG MAKE_DEFINES="-j 8"

# Install build dependencies for OpenSSL
RUN apk update && apk upgrade && \
    apk add build-base linux-headers libtool automake autoconf make git wget

# Clone OpenSSL repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch master https://github.com/openssl/openssl.git

# Build and install OpenSSL
WORKDIR /optbuild/openssl
RUN LDFLAGS="-Wl,-rpath -Wl,${INSTALLDIR_OPENSSL}/lib64" ./config shared --prefix=${INSTALLDIR_OPENSSL} && \
    make ${MAKE_DEFINES} && \
    make install && \
    ln -s ${INSTALLDIR_OPENSSL}/lib64 ${INSTALLDIR_OPENSSL}/lib || true && \
    ln -s ${INSTALLDIR_OPENSSL}/lib ${INSTALLDIR_OPENSSL}/lib64 || true

# Stage 2: Build liboqs
FROM alpine:3.20 AS buildliboqs

# Global build arguments
ARG INSTALLDIR_OPENSSL=/opt/openssl32
ARG INSTALLDIR_LIBOQS=/opt/liboqs
ARG LIBOQS_BUILD_DEFINES="-DOQS_DIST_BUILD=ON"
ARG MAKE_DEFINES="-j 8"
ARG LIBOQS_VERSION="0.10.1"

# Install build dependencies for liboqs
RUN apk add build-base linux-headers libtool automake autoconf cmake ninja make git wget

# Clone liboqs repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch ${LIBOQS_VERSION} https://github.com/open-quantum-safe/liboqs

# Get OpenSSL image (from cache)
COPY --from=buildopenssl ${INSTALLDIR_OPENSSL} ${INSTALLDIR_OPENSSL}

# Build and install liboqs
WORKDIR /optbuild/liboqs
RUN mkdir build && \
    cd build && \
    cmake -G"Ninja" .. -DOPENSSL_ROOT_DIR=${INSTALLDIR_OPENSSL} ${LIBOQS_BUILD_DEFINES} -DCMAKE_INSTALL_PREFIX=${INSTALLDIR_LIBOQS} && \
    ninja install

# Stage 3: Build oqs-provider
FROM alpine:3.20 AS buildoqsprovider

# Global build arguments
ARG INSTALLDIR_OPENSSL=/opt/openssl32
ARG INSTALLDIR_LIBOQS=/opt/liboqs
ARG LIBOQS_BUILD_DEFINES="-DOQS_DIST_BUILD=ON"
ARG MAKE_DEFINES="-j 8"
ARG OQS_PROVIDER_VERSION="0.6.1"

# Install build dependencies for oqs-provider
RUN apk add build-base linux-headers libtool cmake ninja git wget

# Clone oqs-provider repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch ${OQS_PROVIDER_VERSION} https://github.com/open-quantum-safe/oqs-provider.git

# Get openssl32 and liboqs
COPY --from=buildopenssl ${INSTALLDIR_OPENSSL} ${INSTALLDIR_OPENSSL}
COPY --from=buildliboqs ${INSTALLDIR_LIBOQS} ${INSTALLDIR_LIBOQS}

# Build and install oqs-provider
WORKDIR /optbuild/oqs-provider
RUN liboqs_DIR=${INSTALLDIR_LIBOQS} cmake -DOPENSSL_ROOT_DIR=${INSTALLDIR_OPENSSL} -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=${INSTALLDIR_OPENSSL} -S . -B _build && \
    cmake --build _build && \
    cmake --install _build && \
    cp _build/lib/oqsprovider.so ${INSTALLDIR_OPENSSL}/lib64/ossl-modules && \
    sed -i "s/default = default_sect/default = default_sect\noqsprovider = oqsprovider_sect/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf && \
    sed -i "s/\[default_sect\]/\[default_sect\]\nactivate = 1\n\[oqsprovider_sect\]\nactivate = 1\n/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf && \
    sed -i "s/providers = provider_sect/providers = provider_sect\nssl_conf = ssl_sect\n\n\[ssl_sect\]\nsystem_default = system_default_sect\n\n\[system_default_sect\]\nGroups = \$ENV\:\:DEFAULT_GROUPS\n/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf && \
    sed -i "s/HOME\t\t\t= ./HOME           = .\nDEFAULT_GROUPS = kyber768/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf

# Stage 4: Build operator
FROM golang:1.23-alpine AS buildoperator

# Global build arguments
ARG LIBOQS_VERSION="0.10.1"

# Set working directory
WORKDIR /home/qubesec

# Install build dependencies
RUN apk --no-cache add build-base cmake openssl-dev git

# Clone liboqs repository
RUN git clone --depth 1 --branch ${LIBOQS_VERSION} https://github.com/open-quantum-safe/liboqs

# Install liboqs
RUN cmake -S liboqs -B liboqs/build -DBUILD_SHARED_LIBS=ON && \
    cmake --build liboqs/build --parallel 4 && \
    cmake --build liboqs/build --target install

RUN git clone --depth=1 https://github.com/open-quantum-safe/liboqs-go /root/liboqs-go

# Configure liboqs-go
ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/root/liboqs-go/.config
ENV LD_LIBRARY_PATH=/usr/local/lib

# Copy Go source code
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY api/ api/
COPY internal/ internal/

# Build the Go application
RUN go build -o /manager cmd/main.go

# Stage 5: Final image
FROM alpine:3.20

# Global build arguments
ARG INSTALLDIR_OPENSSL=/opt/openssl32

# Copy oqs-provider and operator artifacts
COPY --from=buildoqsprovider ${INSTALLDIR_OPENSSL} ${INSTALLDIR_OPENSSL}
COPY --from=buildoperator /manager /manager
COPY --from=buildoperator /usr/local/lib /usr/local/lib

# Set environment variables
ENV PATH="${INSTALLDIR_OPENSSL}/bin:${PATH}"

# Set user
USER 65532:65532

# Set entry point
ENTRYPOINT ["/manager"]
