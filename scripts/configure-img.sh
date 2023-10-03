#!/bin/bash

# This script should be run as root

# Ask for the new username
read -p "Enter the new username: " username

# Create the user
adduser $username

# Grant sudo privileges
usermod -aG sudo $username

# Set up SSH key-based authentication
echo "Setting up SSH key-based authentication..."

# Create .ssh directory for the new user
mkdir /home/$username/.ssh
chmod 700 /home/$username/.ssh

# Ask for the public key
read -p "Paste the public SSH key for the new user: " ssh_key

# Add the public key to authorized_keys
echo $ssh_key >> /home/$username/.ssh/authorized_keys
chmod 600 /home/$username/.ssh/authorized_keys
chown $username:$username /home/$username/.ssh -R

# Disable root SSH access
sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config

# Restart SSH
service ssh restart

echo "User setup complete. Root SSH access has been disabled."
