#!/bin/bash
# OmniLang Installation Script

set -e

INSTALL_DIR="/usr/local/omni"
BIN_DIR="/usr/local/bin"

echo "Installing OmniLang {{VERSION}}..."

# Create installation directory
sudo mkdir -p $INSTALL_DIR
sudo mkdir -p $BIN_DIR

# Copy files
sudo cp -r . $INSTALL_DIR/

# Create symlinks
sudo ln -sf $INSTALL_DIR/omnic $BIN_DIR/omnic
sudo ln -sf $INSTALL_DIR/omnir $BIN_DIR/omnir

# Set permissions
sudo chmod +x $INSTALL_DIR/omnic
sudo chmod +x $INSTALL_DIR/omnir

echo "OmniLang installed successfully!"
echo "Run 'omnic --version' to verify installation"

