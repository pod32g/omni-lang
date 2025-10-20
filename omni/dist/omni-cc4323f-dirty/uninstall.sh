#!/bin/bash
# OmniLang Uninstallation Script

set -e

INSTALL_DIR="/usr/local/omni"
BIN_DIR="/usr/local/bin"

echo "Uninstalling OmniLang..."

# Remove symlinks
sudo rm -f $BIN_DIR/omnic
sudo rm -f $BIN_DIR/omnir

# Remove installation directory
sudo rm -rf $INSTALL_DIR

echo "OmniLang uninstalled successfully!"
