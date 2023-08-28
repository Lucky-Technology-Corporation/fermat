#!/bin/bash

# Set DEBIAN_FRONTEND to noninteractive and use yes to automatically say "yes" to prompts
export DEBIAN_FRONTEND=noninteractive

# Added in Ubuntu 22.04 and causes issues with scripts since we are not running in an interactive mode
# remove package; see more here: https://askubuntu.com/questions/1367139/apt-get-upgrade-auto-restart-services
sudo DEBIAN_FRONTEND=noninteractive apt-get remove needrestart -y -qq

# Update the apt package list
sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq

sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq ca-certificates curl gnupg

yes | sudo install -m 0755 -d /etc/apt/keyrings

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

yes | sudo chmod a+r /etc/apt/keyrings/docker.gpg

echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" |
  sudo tee /etc/apt/sources.list.d/docker.list >/dev/null


sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq

sudo DEBIAN_FRONTEND=noninteractive apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin -y -qq

sudo groupadd docker

newgrp docker

# Allow current user to run Docker commands
sudo usermod -aG docker $USER

sudo chown "$USER":"$USER" /home/"$USER"/.docker -R
sudo chmod g+rwx "$HOME/.docker" -R

sudo systemctl enable docker.service
sudo systemctl enable containerd.service

mkdir -p ~/.docker/cli-plugins/
curl -SL https://github.com/docker/compose/releases/download/v2.3.3/docker-compose-linux-x86_64 -o ~/.docker/cli-plugins/docker-compose

# Set executable permissions
chmod +x ~/.docker/cli-plugins/docker-compose

# Install UFW
#sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq ufw

# Set default policies to deny all incoming and outgoing connections
#sudo ufw default deny incoming
#sudo ufw default allow outgoing
#
## Allow SSH on port 2222
#sudo ufw allow 2222/tcp
#
## Allow external MongoDB access (default MongoDB port is 27017)
#sudo ufw allow 27017/tcp
#
## Enable UFW and automatically answer "yes" to the prompt
#yes | sudo ufw enable

# Ensure Docker is running as a process: allows service to restart even if underlying
# VM stops and starts
sudo systemctl enable docker

# Print Docker and Docker Compose versions
docker --version
docker compose --version

# Configure certbot
sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install certbot -y -qq

exit
