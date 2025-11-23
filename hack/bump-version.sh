#!/bin/sh
# hack/bump-version.sh
# Usage: ./hack/bump-version.sh <new_version>
# Bumps manifest versions in .version

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
sed_replace "s/v[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}/v$NEW_VERSION/g" .version

echo "Bumped manifest versions to v$NEW_VERSION in docs."