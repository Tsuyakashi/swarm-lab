#!/bin/bash
(set -o posix; [ -f /usr/bin/dos2unix ] || (sudo apt-get update &>/dev/null && sudo apt-get install -y dos2unix &>/dev/null)) && dos2unix "$0"

# Install Docker
echo "Installing docker & docker compose plugin"
if ! command -v docker &> /dev/null || ! docker compose version &> /dev/null; then
    sudo DEBIAN_FRONTEND=noninteractive apt-get update &>/dev/null && \
        sudo DEBIAN_FRONTEND=noninteractive apt-get install -y docker.io docker-compose-v2 &>/dev/null
fi

sudo usermod -aG docker vagrant

sudo docker info 2>/dev/null | grep -q "Swarm: active" \
    || sudo docker swarm init --advertise-addr "${MANAGER_IP}" >/dev/null

APP_DIR="/tmp"
cd "$APP_DIR" || exit 1

echo "Pulling images"
sudo -E docker compose pull &>/dev/null

echo "Deploying stack"
sudo -E docker stack deploy --with-registry-auth -c docker-compose.yml stage

echo "All done"
