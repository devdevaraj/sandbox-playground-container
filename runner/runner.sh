#!/bin/sh

json_file="/resourses/.configs/$1.json"

number=$(jq -r '.vms // error("field \"vms\" missing or null")' "$json_file")

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

EXECUTABLE="$SCRIPT_DIR/firestarter"
WSS_EXECUTABLE="$SCRIPT_DIR/ws-server"
FS_DIR="/root/firecracker/overlayfs"

mkdir -p "$FS_DIR"
for  i in $(seq 1 "$number"); do
dd if=/dev/zero of="$FS_DIR/vm$i-overlay.ext4" conv=sparse bs=1M count=40960 && mkfs.ext4 "$FS_DIR/vm$i-overlay.ext4"
done

"$WSS_EXECUTABLE" "$1" &

until "$EXECUTABLE" "$1"; do
  echo "Command failed. Retrying in 1 seconds..."
  sleep 1
done


echo "Command succeeded."
