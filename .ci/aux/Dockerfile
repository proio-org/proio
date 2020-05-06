# Description: An image use in ProIO testing

FROM archlinux/base

ARG BUILD_THREADS=5

ARG PROTOBUF_VERSION=latest
ARG LZ4_VERSION=latest

RUN pacman -Sy --noconfirm \
        base-devel \
        cmake \
        git \
        libffi \
        unzip \
&&  rm -rf /var/cache/pacman/pkg

RUN git clone https://github.com/lz4/lz4.git \
&&  cd lz4 \
&&  if [ "$LZ4_VERSION" != "latest" ]; then git checkout tags/$LZ4_VERSION; fi \
&&  make -j$BUILD_THREADS install \
&&  cd ../ \
&&  rm -rf lz4

RUN git clone https://github.com/google/protobuf.git \
&&  cd protobuf \
&&  if [ "$PROTOBUF_VERSION" != "latest" ]; then git checkout tags/$PROTOBUF_VERSION; else git checkout $(git describe --abbrev=0 --tags); fi \
&&  git submodule update --init \
&&  bash autogen.sh \
&&  ./configure \
&&  make -j$BUILD_THREADS install \
&&  cd ../ \
&&  rm -rf protobuf
