#!/bin/sh
set -e
go build -o loom .
cp loom ~/.local/bin/loom
echo "loom built and installed to ~/.local/bin/loom"
