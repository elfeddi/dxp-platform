#!/bin/bash
export NVM_DIR="/home/azureuser/.nvm"
source "$NVM_DIR/nvm.sh"
nvm use 20
cd /home/azureuser/dxp-platform/dxp-portal
exec yarn workspace backend start
