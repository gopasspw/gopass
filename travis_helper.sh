#!/usr/bin/env bash

set -eux

if [ $TRAVIS_OS_NAME = linux ]; then
  make $1
else
  make $1-$TRAVIS_OS_NAME
fi
