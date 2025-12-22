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

# Replace all occurrences of vX.Y.Z with the new version (vNEW_VERSION)
sed -Ei "s/v[0-9]+\.[0-9]+\.[0-9]+(\-[a-zA-Z0-9\-\.]+)*/v$NEW_VERSION/g" .version
sed -Ei "s/v[0-9]+\.[0-9]+\.[0-9]+(\-[a-zA-Z0-9\-\.]+)*\/install\.yaml/v$NEW_VERSION\/install.yaml/g" doc/docs/get-started.md
sed -Ei "s/gate\-operator version v[0-9]+\.[0-9]+\.[0-9]+(\-[a-zA-Z0-9\-\.]+)*/gate-operator version v$NEW_VERSION/g" cmd/main.go

echo "Bumped manifest versions to v$NEW_VERSION in docs."