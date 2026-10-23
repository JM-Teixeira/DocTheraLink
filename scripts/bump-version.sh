#!/usr/bin/env bash
# Sets the Klyniva version (major.minor.patch) in the backend and frontend.
# Usage: scripts/bump-version.sh 0.7.0
set -euo pipefail
v="${1:?usage: $0 <major.minor.patch>}"
[[ "$v" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "not a version: $v" >&2; exit 1; }
root="$(cd "$(dirname "$0")/.." && pwd)"
sed -i -E "s/^const Version = \".*\"/const Version = \"$v\"/" "$root/be/internal/version/version.go"
sed -i -E "0,/\"version\": \"[^\"]*\"/s//\"version\": \"$v\"/" "$root/fe/package.json"
echo "Klyniva $v"
