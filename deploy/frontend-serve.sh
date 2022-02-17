#!/bin/bash

# Intended for use in a docker container or an init container for a kubernetes
# pod. This idempotently creates a yarn build depending on changes of the
# CAPTCHA_SITE_KEY and contents of the frontend source directory
set -euxo pipefail

src_dir="$1"
dest_dir="$2"
site_key="${CAPTCHA_SITE_KEY}"

src_base_dir=$(dirname "$src_dir")
src_base_name=$(basename "$src_dir")

mkdir -p "$(dirname "$dest_dir")"

build_sha_path="${dest_dir}/build.shasum"

stored_build_sha=""
if [ -f "$build_sha_path" ]; then
    stored_build_sha=$(cat "$build_sha_path")
fi

# Get the source sha by tarring all files and getting the shasum of that. For a
# given docker container this would be deterministic.
src_sha=$(tar -C "${src_base_dir}" -cf - --sort=name --mtime='1970-01-01' --exclude=".parcel-cache" "${src_base_name}" | sha256sum | awk '{print $1}')

# Add the site key to the hash
src_build_sha=$(echo "${src_sha}${site_key}" | sha256sum | awk '{print $1}')

# If there's nothing to do then we don't need to rebuild the frontend
if [ "$src_build_sha" = "$stored_build_sha" ]; then
    echo "No changes"
    exit 0
fi

# Build
cd "$src_dir"

export CAPTCHA_SITE_KEY="${site_key}"
yarn build

mv dist "${dest_dir}"

# Store the build sha
echo "${src_build_sha}" > "${build_sha_path}"
