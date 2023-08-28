#!/bin/bash

# Create a new user
NEW_USERNAME="swizzle_prod_user" # Adjust the name as needed
adduser --disabled-password --gecos "" $NEW_USERNAME

# Add the user to the sudoers
echo "$NEW_USERNAME ALL=(ALL) NOPASSWD:ALL" | tee -a /etc/sudoers

# Copy the root's SSH key to the new user's .ssh directory
mkdir -p /home/$NEW_USERNAME/.ssh
cp /root/.ssh/authorized_keys /home/$NEW_USERNAME/.ssh/
chown -R $NEW_USERNAME:$NEW_USERNAME /home/$NEW_USERNAME/.ssh
chmod 700 /home/$NEW_USERNAME/.ssh
chmod 600 /home/$NEW_USERNAME/.ssh/authorized_keys

# Disable root SSH access
sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config

# Change the SSH port (e.g., to 2222)
NEW_SSH_PORT=2222
sed -i "s/#Port 22/Port $NEW_SSH_PORT/" /etc/ssh/sshd_config

# Disable all authentication methods except SSH key-based authentication
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
echo "ChallengeResponseAuthentication no" >> /etc/ssh/sshd_config
echo "UsePAM no" >> /etc/ssh/sshd_config

# Restart the SSH service for changes to take effect
service ssh restart

echo "VM Init Successful!"

exit
