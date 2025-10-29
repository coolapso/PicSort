#!/bin/bash

## check for root privileges
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

# Variables
REPO="coolapso/picsort"
VERSION=${VERSION:-"latest"}
INSTALL_DIR="/usr/local/bin"
DESKTOP_DIR="/usr/share/applications"
ICON_DIR="/usr/share/pixmaps"

# Determine OS and Architecture
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
elif [[ "$ARCH" == "aarch64" || "$ARCH" == "arm64" ]]; then
  ARCH="arm64"
elif [[ "$ARCH" == "i386" || "$ARCH" == "i686" ]]; then
  ARCH="386"
fi

# Fetch the latest release if no version is specified
if [[ "$VERSION" == "latest" ]]; then
  VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

# Download URL for the release file
FILE="picsort_${VERSION:1}_${OS}_${ARCH}.tar.gz"
FILE_URL="https://github.com/$REPO/releases/download/$VERSION/$FILE"

if ! curl -LO "$FILE_URL"; then
  echo "Failed to download $FILE_URL"
  exit 1
fi

# Create a temporary install dir
if [[ ! -d picsort-install ]]; then
  mkdir picsort-install
fi
mv "$FILE" picsort-install
cd picsort-install || exit

# Extract and install
if ! tar xzf "$FILE"; then
  echo "Failed to extract $FILE"
  exit 1
fi

if ! chmod +x picsort; then
  echo "Failed to make picsort executable"
  exit 1
fi

if ! mv picsort "$INSTALL_DIR"; then
  echo "Failed to move picsort to $INSTALL_DIR"
  exit 1
fi

## Not installing picsort via package manager, update the
## desktop file to execute the binary from /usr/local/bin
sed -i 's|usr/bin/|usr/local/bin/|' build/picsort.desktop

if ! mv build/picsort.desktop "$DESKTOP_DIR"; then
  echo "Failed to move picsort desktop file to $DESKTOP_DIR"
  exit 1
fi

if ! mv media/logo.png "$ICON_DIR/picsort.png"; then
  echo "Failed to move logo.png to $ICON_DIR"
  exit 1
fi

# Cleanup
cd ../ || exit
rm -rf picsort-install

echo "Installation of $REPO version $VERSION complete."
