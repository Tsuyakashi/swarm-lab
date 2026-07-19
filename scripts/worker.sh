#!/bin/bash
(set -o posix; [ -f /usr/bin/dos2unix ] || (sudo apt-get update &>/dev/null && sudo apt-get install -y dos2unix &>/dev/null)) && dos2unix "$0"

# Install Docker
echo "Installing docker"
if ! command -v docker &> /dev/null; then
    sudo DEBIAN_FRONTEND=noninteractive apt-get update &>/dev/null \
        && sudo DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io &>/dev/null
fi

sudo usermod -aG docker vagrant

echo "All done"
