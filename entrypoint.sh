#!/bin/sh
set -eu

VERSION=$(/usr/local/bin/git-version)
echo "version=$VERSION" >> "$GITHUB_OUTPUT"
echo "$VERSION"
