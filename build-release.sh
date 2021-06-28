#!/usr/bin/env bash

set -e

goreleaser build --rm-dist

echo "Output distribution found in dist"
