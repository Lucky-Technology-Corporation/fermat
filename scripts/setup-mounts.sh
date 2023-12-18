#!/bin/bash

if [ -z "$MONGO_VOLUME_NAME" ]; then
    echo "MONGO_VOLUME_NAME variable must be set."
    exit 1
fi

mongo_dev_name=$(lsblk -o NAME,LABEL | grep mongodata | awk '{print $1}')
if [ -z "$mongo_dev_name" ]; then
    echo "No mongo device found."
    exit 1
fi

mongo_dev="/dev/$mongo_dev_name"
echo "Found mongo volume device: $mongo_dev"

home_dev_name=$(lsblk -o NAME,LABEL | grep swizzlehome | awk '{print $1}')
if [ -z "$home_dev_name" ]; then
    echo "No home device found."
    exit 1
fi

home_dev="/dev/$home_dev_name"

mongo_mnt=$(findmnt -n -o TARGET --source $mongo_dev)
echo "Checking if $mongo_dev is mounted"

if [ -z "$mongo_mnt" ]; then
    # Mongo is not mounted so we need mount it to /mnt/$MONGO_VOLUME_NAME
    echo "Mongo volume isn't mounted. Mounting to: /mnt/$MONGO_VOLUME_NAME now..."
    mkdir -p /mnt/$MONGO_VOLUME_NAME
    mount $mongo_dev /mnt/$MONGO_VOLUME_NAME
    if [ $? -eq 0 ]; then
        echo "Successfully mounted mongo volume at /mnt/$MONGO_VOLUME_NAME"
    else
        echo "Failed to mount mongo volume"
        exit 1
    fi
fi

home_mnt=$(findmnt -n -o TARGET --source $home_dev)

# If home has not been auto-mounted then we need to give to temporarily mount it.
if [ -z "$home_mnt" ]; then
    home_mnt=/mnt/swizzlehome
    swizzle_home=/home/swizzle
    echo "Home directory is not mounted. Mounting to temporary location at $home_mnt..."

    mkdir -p $home_mnt
    mount $home_dev $home_mnt
    if [ $? -eq 0 ]; then
        echo "Successfully mounted home directory at $home_mnt"
    else
        echo "Failed to mount home directory"
        exit 1
    fi
fi

echo "Syncing current home directory to remote..."
rsync -av $swizzle_home/ $home_mnt/
if [ $? -eq 0 ]; then
    echo "Synced! Now deleting all files in the home directory..."
else
    echo "Failed to sync current home directory to remote"
    exit 1
fi

# If swizzle_home is not set this will delete the whole droplet!
# Sanity check (even though we set it manually above).
if [ -z "$swizzle_home" ]; then
    echo "swizzle_home variable isn't set properly."
    exit 1
fi
rm -rf "$swizzle_home/*"

echo "Unmounting home directory and re-mounting at $swizzle_home"
umount -f $home_dev
if [ $? -eq 0 ]; then
    echo "Successfully unmounted the home directory. Now re-mounting to $swizzle_home"
else
    echo "Failed to unmount home directory"
    exit 1
fi

mount $home_dev $swizzle_home
if [ $? -eq 0 ]; then
    echo "Successfully mounted the home directory!"
else
    echo "Failed to mount home directory"
    exit 1
fi
