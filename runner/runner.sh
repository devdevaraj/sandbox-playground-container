#!/bin/bash

echo "$2"

json_file="/resourses/.configs/$1.json"

number=$(jq -r '.templates | length' "$json_file")

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

EXECUTABLE="$SCRIPT_DIR/firestarter"
WSS_EXECUTABLE="$SCRIPT_DIR/ws-server"
FS_DIR="/root/firecracker/overlayfs"

mkdir -p "$FS_DIR"
for  i in $(seq 0 $((number - 1))); do
 enableOverlay=$(jq -r ".templates[$i][\"enable-overlay\"]" "$json_file")
 if [[ $enableOverlay == "true" ]]; then
  overlaySize=$(jq -r ".templates[$i][\"overlay-size\"]" "$json_file")
  vm_number=$((i + 1))
  dd if=/dev/zero of="$FS_DIR/vm$vm_number-overlay.ext4" conv=sparse bs=1M count=$overlaySize && mkfs.ext4 "$FS_DIR/vm$vm_number-overlay.ext4"
 fi
done

"$WSS_EXECUTABLE" "$1" &

until "$EXECUTABLE" "$1" "/resourses/zfs/$2/rootfs_master.ext4"; do
  echo "Command failed. Retrying in 1 seconds..."
  sleep 1
done


echo "Command succeeded."