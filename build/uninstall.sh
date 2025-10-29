#!/bin/bash

## check for root privileges
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit
fi

if ! rm -rf /usr/local/bin/picsort; then
  echo "Failed to remove picsort from /usr/local/bin"
  exit 1
fi

if ! rm -rf /usr/share/applications/picsort.desktop; then
  echo "Failed to remove picsort desktop file from /usr/share/applications"
fi

if ! rm -rf /usr/share/pixmaps/picsort.png; then
  echo "Failed to remove picsort logo from /usr/share/pixmaps"
fi

echo "Picsort uninstalled successfully, thank you for using it!"
