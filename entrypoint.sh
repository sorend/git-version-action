#!/bin/sh
set -eu

/usr/local/bin/git-version > "$GITHUB_OUTPUT"
