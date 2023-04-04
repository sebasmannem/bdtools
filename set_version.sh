#!/bin/bash
VERSION=$(echo $GITHUB_REF | cut -d / -f 3)
if [ -z "${VERSION}" ]; then
  VERSION=$(git tag | sort -V | grep '^v' | tail -n1)-devel
fi
echo "package internal

const appVersion = \"$VERSION\"" > internal/version.go
