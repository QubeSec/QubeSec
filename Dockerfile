# Global build arguments
ARG BASE_ALPINE_VERSION=3.20
ARG BASE_GOLANG_VERSION=1.23
ARG LIBOQS_VERSION="0.10.1"
ARG LIBOQS_GO_VERSION="0.10.0"
ARG OQS_PROVIDER_VERSION="0.6.1"
ARG OPENSSL_VERSION="3.3.2"
ARG OPENSSL_INSTALLDIR=/opt/openssl
ARG LIBOQS_INSTALLDIR=/opt/liboqs

# Stage 1: Build OpenSSL
FROM alpine:${BASE_ALPINE_VERSION} AS buildopenssl

# Re-declare the global argument for this stage
ARG OPENSSL_VERSION
ARG OPENSSL_INSTALLDIR

# Install build dependencies for OpenSSL
RUN apk add build-base linux-headers libtool automake autoconf make git wget

# Clone OpenSSL repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch openssl-${OPENSSL_VERSION} https://github.com/openssl/openssl.git

# Build and install OpenSSL
WORKDIR /optbuild/openssl
RUN LDFLAGS="-Wl,-rpath -Wl,${OPENSSL_INSTALLDIR}/lib64" ./config shared --prefix=${OPENSSL_INSTALLDIR} && \
    make -j 8 && \
    make install && \
    ln -s ${OPENSSL_INSTALLDIR}/lib64 ${OPENSSL_INSTALLDIR}/lib || true && \
    ln -s ${OPENSSL_INSTALLDIR}/lib ${OPENSSL_INSTALLDIR}/lib64 || true

# Stage 2: Build liboqs
ARG BASE_ALPINE_VERSION
FROM alpine:${BASE_ALPINE_VERSION} AS buildliboqs

# Re-declare the global argument for this stage
ARG LIBOQS_VERSION
ARG LIBOQS_INSTALLDIR
ARG OPENSSL_INSTALLDIR

# Install build dependencies for liboqs
RUN apk add build-base linux-headers libtool automake autoconf cmake ninja make git wget

# Clone liboqs repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch ${LIBOQS_VERSION} https://github.com/open-quantum-safe/liboqs.git

# Get OpenSSL image (from cache)
COPY --from=buildopenssl ${OPENSSL_INSTALLDIR} ${OPENSSL_INSTALLDIR}

# Build and install liboqs
WORKDIR /optbuild/liboqs
RUN mkdir build && \
    cd build && \
    cmake -G"Ninja" .. -DOPENSSL_ROOT_DIR=${OPENSSL_INSTALLDIR} -DOQS_DIST_BUILD=ON -DCMAKE_INSTALL_PREFIX=${LIBOQS_INSTALLDIR} && \
    ninja install

# Stage 3: Build oqs-provider
ARG BASE_ALPINE_VERSION
FROM alpine:${BASE_ALPINE_VERSION} AS buildoqsprovider

# Re-declare the global argument for this stage
ARG OQS_PROVIDER_VERSION
ARG LIBOQS_INSTALLDIR
ARG OPENSSL_INSTALLDIR

# Install build dependencies for oqs-provider
RUN apk add build-base linux-headers libtool cmake ninja git wget

# Clone oqs-provider repository
RUN mkdir /optbuild && \
    cd /optbuild && \
    git clone --depth 1 --branch ${OQS_PROVIDER_VERSION} https://github.com/open-quantum-safe/oqs-provider.git

# Get openssl32 and liboqs
COPY --from=buildopenssl ${OPENSSL_INSTALLDIR} ${OPENSSL_INSTALLDIR}
COPY --from=buildliboqs ${LIBOQS_INSTALLDIR} ${LIBOQS_INSTALLDIR}

# Build and install oqs-provider
WORKDIR /optbuild/oqs-provider
RUN liboqs_DIR=${LIBOQS_INSTALLDIR} cmake -DOPENSSL_ROOT_DIR=${OPENSSL_INSTALLDIR} -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=${OPENSSL_INSTALLDIR} -S . -B _build && \
    cmake --build _build && \
    cmake --install _build && \
    cp _build/lib/oqsprovider.so ${OPENSSL_INSTALLDIR}/lib64/ossl-modules && \
    sed -i "s/default = default_sect/default = default_sect\noqsprovider = oqsprovider_sect/g" ${OPENSSL_INSTALLDIR}/ssl/openssl.cnf && \
    sed -i "s/\[default_sect\]/\[default_sect\]\nactivate = 1\n\[oqsprovider_sect\]\nactivate = 1\n/g" ${OPENSSL_INSTALLDIR}/ssl/openssl.cnf && \
    sed -i "s/providers = provider_sect/providers = provider_sect\nssl_conf = ssl_sect\n\n\[ssl_sect\]\nsystem_default = system_default_sect\n\n\[system_default_sect\]\nGroups = \$ENV\:\:DEFAULT_GROUPS\n/g" ${OPENSSL_INSTALLDIR}/ssl/openssl.cnf && \
    sed -i "s/HOME\t\t\t= ./HOME           = .\nDEFAULT_GROUPS = kyber768/g" ${OPENSSL_INSTALLDIR}/ssl/openssl.cnf

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
RUN git clone --depth 1 --branch ${LIBOQS_VERSION} https://github.com/open-quantum-safe/liboqs

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

# Re-declare the global argument for this stage
ARG OPENSSL_INSTALLDIR

# Copy oqs-provider and operator artifacts
COPY --from=buildoqsprovider ${OPENSSL_INSTALLDIR} ${OPENSSL_INSTALLDIR}
COPY --from=buildoperator /manager /manager
COPY --from=buildoperator /usr/local/lib /usr/local/lib

# Set environment variables
ENV PATH="${OPENSSL_INSTALLDIR}/bin:${PATH}"

# Set user
USER 65532:65532

# Set entry point
ENTRYPOINT ["/manager"]
