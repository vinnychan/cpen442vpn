#!/bin/sh
is_installed(){
  echo "Checking if $1 is installed."
  local status="$1 is installed."
  type $1>/dev/null 2>&1 ||
    { local status="$1 is not installed, Aborting"; exit 1;}
  echo "$status";
  echo "---"
}

is_installed go
is_installed node
is_installed electron

echo "All pre-requisites are installed."
echo "Starting VPN program..."

electron gui/main.js
