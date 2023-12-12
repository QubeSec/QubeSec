# Multi-stage build: First the full builder image:

ARG INSTALLDIR_OPENSSL=/opt/openssl32
ARG INSTALLDIR_LIBOQS=/opt/liboqs

# liboqs build type variant; maximum portability of image:
ARG LIBOQS_BUILD_DEFINES="-DOQS_DIST_BUILD=ON"

# Define the degree of parallelism when building the image; leave the number away only if you know what you are doing
ARG MAKE_DEFINES="-j 8"

ARG SIG_ALG="dilithium3"

FROM alpine:3.19 as buildopenssl
# Take in all global args
ARG INSTALLDIR_OPENSSL
ARG INSTALLDIR_LIBOQS
ARG LIBOQS_BUILD_DEFINES
ARG MAKE_DEFINES
ARG SIG_ALG

LABEL version="1"
ENV DEBIAN_FRONTEND noninteractive

RUN apk update && apk upgrade

# Get all software packages required for builing openssl
RUN apk add build-base linux-headers \
            libtool automake autoconf \
            make \
            git wget

# get current openssl sources
RUN mkdir /optbuild && cd /optbuild && git clone --depth 1 --branch master https://github.com/openssl/openssl.git

# build OpenSSL3
WORKDIR /optbuild/openssl
RUN LDFLAGS="-Wl,-rpath -Wl,${INSTALLDIR_OPENSSL}/lib64" ./config shared --prefix=${INSTALLDIR_OPENSSL} && \
    make ${MAKE_DEFINES} && make install && if [ -d ${INSTALLDIR_OPENSSL}/lib64 ]; then ln -s ${INSTALLDIR_OPENSSL}/lib64 ${INSTALLDIR_OPENSSL}/lib; fi && if [ -d ${INSTALLDIR_OPENSSL}/lib ]; then ln -s ${INSTALLDIR_OPENSSL}/lib ${INSTALLDIR_OPENSSL}/lib64; fi 

FROM alpine:3.19 as buildliboqs
# Take in all global args
ARG INSTALLDIR_OPENSSL
ARG INSTALLDIR_LIBOQS
ARG LIBOQS_BUILD_DEFINES
ARG MAKE_DEFINES
ARG SIG_ALG

LABEL version="1"
ENV DEBIAN_FRONTEND noninteractive

# Get all software packages required for builing liboqs:
RUN apk add build-base linux-headers \
            libtool automake autoconf cmake ninja \
            make \
            git wget

# Get OpenSSL image (from cache)
COPY --from=buildopenssl ${INSTALLDIR_OPENSSL} ${INSTALLDIR_OPENSSL}

RUN mkdir /optbuild && cd /optbuild && git clone --depth 1 --branch main https://github.com/open-quantum-safe/liboqs

WORKDIR /optbuild/liboqs
RUN mkdir build && cd build && cmake -G"Ninja" .. -DOPENSSL_ROOT_DIR=${INSTALLDIR_OPENSSL} ${LIBOQS_BUILD_DEFINES} -DCMAKE_INSTALL_PREFIX=${INSTALLDIR_LIBOQS} && ninja install

FROM alpine:3.19 as buildoqsprovider
# Take in all global args
ARG INSTALLDIR_OPENSSL
ARG INSTALLDIR_LIBOQS
ARG LIBOQS_BUILD_DEFINES
ARG MAKE_DEFINES
ARG SIG_ALG

LABEL version="1"
ENV DEBIAN_FRONTEND noninteractive

# Get all software packages required for builing oqsprovider
RUN apk add build-base linux-headers \
            libtool cmake ninja \
            git wget

RUN mkdir /optbuild && cd /optbuild && git clone --depth 1 --branch main https://github.com/open-quantum-safe/oqs-provider.git

# Get openssl32 and liboqs
COPY --from=buildopenssl ${INSTALLDIR_OPENSSL} ${INSTALLDIR_OPENSSL}
COPY --from=buildliboqs ${INSTALLDIR_LIBOQS} ${INSTALLDIR_LIBOQS}

# build & install provider (and activate by default)
WORKDIR /optbuild/oqs-provider
RUN liboqs_DIR=${INSTALLDIR_LIBOQS} cmake -DOPENSSL_ROOT_DIR=${INSTALLDIR_OPENSSL} -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH=${INSTALLDIR_OPENSSL} -S . -B _build && cmake --build _build  && cmake --install _build && cp _build/lib/oqsprovider.so ${INSTALLDIR_OPENSSL}/lib64/ossl-modules && sed -i "s/default = default_sect/default = default_sect\noqsprovider = oqsprovider_sect/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf && sed -i "s/\[default_sect\]/\[default_sect\]\nactivate = 1\n\[oqsprovider_sect\]\nactivate = 1\n/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf && sed -i "s/providers = provider_sect/providers = provider_sect\nssl_conf = ssl_sect\n\n\[ssl_sect\]\nsystem_default = system_default_sect\n\n\[system_default_sect\]\nGroups = \$ENV\:\:DEFAULT_GROUPS\n/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf && sed -i "s/HOME\t\t\t= ./HOME           = .\nDEFAULT_GROUPS = kyber768/g" ${INSTALLDIR_OPENSSL}/ssl/openssl.cnf

# Use the Alpine version of the Golang base image
FROM golang:1.21-alpine as buildoperator

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

FROM alpine:3.19

ARG INSTALLDIR_OPENSSL

COPY --from=buildoqsprovider ${INSTALLDIR_OPENSSL} ${INSTALLDIR_OPENSSL}
COPY --from=buildoperator /manager /manager
COPY --from=buildoperator /usr/local/lib /usr/local/lib

ENV PATH="${INSTALLDIR_OPENSSL}/bin:${PATH}"

USER 65532:65532

# Set the entry point for the container
ENTRYPOINT ["/manager"]
