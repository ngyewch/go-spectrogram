#!/usr/bin/env bash

set -e

goreleaser build --snapshot --rm-dist

echo "Output distribution found in dist"
