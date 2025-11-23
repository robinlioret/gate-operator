#!/bin/sh
# hack/chart-bump-version.sh
# Usage: ./hack/chart-bump-version.sh <new_version>
# Bumps chart versions in .version-chart, doc/docs/get-started.md

set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <new_version>"
  exit 1
fi

NEW_VERSION="$1"

# Portable sed replacement: write to temp file, then move back
sed_replace() {
  local pattern="$1"
  local file="$2"
  local tmpfile="${file}.tmp.$$"
  sed "$pattern" "$file" > "$tmpfile" && mv "$tmpfile" "$file"
}

# Replace all occurrences of vX.Y.Z with the new version (vNEW_VERSION)
sed_replace "s/[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}/$NEW_VERSION/g" .version-chart
sed_replace "s/oci:\/\/ghcr.io\/robinlioret\/gate-operator\/gate-operator:[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}/oci://ghcr.io/robinlioret/gate-operator/gate-operator:$NEW_VERSION/g" ./doc/docs/get-started.md

echo "Bumped chart versions to $NEW_VERSION in docs."