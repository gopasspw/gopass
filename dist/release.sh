#!/bin/bash
set -e

COMMIT=$(git rev-parse --short=8 HEAD)
VERSION=$(cat VERSION)
DIR=$(pwd)

RELDIR=${PWD}/releases/gopass/${VERSION}

# Prepare completion
make completion

# Clean up
make clean
rm -rf "${RELDIR:?}/"
mkdir -p "${RELDIR}"

# Create source tarball
echo "Creating source tarball ..." \
  && rm -f "/tmp/gopass-${VERSION}.tar.gz" \
  && rm -rf "/tmp/gopass-${VERSION}/" \
  && rsync --exclude=".git" --exclude="releases/" -a . "/tmp/gopass-${VERSION}/" \
  && cd /tmp \
  && echo "${COMMIT}" >"/tmp/gopass-${VERSION}/COMMIT" \
  && tar -czf "gopass-${VERSION}.tar.gz" "gopass-${VERSION}/" \
  && mv "/tmp/gopass-${VERSION}.tar.gz" "${RELDIR}/gopass-${VERSION}.tar.gz" \
  && rm -rf "/tmp/gopass-${VERSION}" \
  && echo "Created source tarball ${RELDIR}/gopass-${VERSION}.tar.gz"

cd "$DIR"

# Cross-compile binaries
for TARGET in \
  darwin/386 \
  darwin/amd64 \
  freebsd/386 \
  freebsd/amd64 \
  freebsd/arm \
  linux/386 \
  linux/amd64 \
  linux/arm \
  linux/arm64 \
  linux/ppc64 \
  linux/ppc64le \
  linux/mips64 \
  linux/mips64le \
  netbsd/386 \
  netbsd/amd64 \
  netbsd/arm \
  openbsd/386 \
  openbsd/amd64 \
; do
  GOOS=$(echo $TARGET | cut -d'/' -f1)
  GOARCH=$(echo $TARGET | cut -d'/' -f2)
  export GOOS GOARCH
  echo "Cross-Compiling for ${GOOS}/${GOARCH}"
  rm -rf "gopass-${VERSION}/"
  make build && \
    mkdir "gopass-${VERSION}" && \
    cp "gopass-${GOOS}-${GOARCH}" "gopass-${VERSION}/gopass" && \
    tar -czf "${RELDIR}/gopass-${VERSION}-${GOOS}-${GOARCH}.tar.gz" "gopass-${VERSION}/" && \
    rm -rf "gopass-${VERSION}"
done

# Build Linux distro packages
for TARGET in \
  deb/386 \
  deb/amd64 \
  rpm/386 \
  rpm/amd64 \
  pacman/386 \
  pacman/amd64 \
; do
  FLAVOR=$(echo $TARGET | cut -d'/' -f1)
  ARCH=$(echo $TARGET | cut -d'/' -f2)
  echo "Building package for ${FLAVOR}/${ARCH}"
  fpm \
    -s dir \
    -t "${FLAVOR}" \
    -a "${ARCH}" \
    -n gopass \
    -v "${VERSION}" \
    -d git \
    -d gnupg \
    --license MIT \
    -m gopass@justwatch.com \
    --url https://www.justwatch.com/gopass \
    -p "${RELDIR}" \
    "gopass-linux-${ARCH}=/usr/bin/gopass"
done

cd "$DIR"
# Generate SHA256SUMS
cd "${RELDIR}" && \
  sha256sum ./* >"gopass_${VERSION}_SHA256SUMS"

cd "$DIR"
# Clean up
make clean
