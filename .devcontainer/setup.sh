#!/bin/bash

# Check if Task is installed in the bin directory and install it if not
if [ ! -f "/workdir/bin/task" ]; then
  echo "Task not found. Installing task..."
  cd /workdir && ./tools/bin/install-dev-tools.bash
  echo "Task installed."
else
  echo "Task is already installed."
fi

# Display some helpful information
echo
echo "FerretDB Development Environment"
echo "================================"
echo
echo "Available commands:"
echo "  task -l                   # List all available tasks"
echo "  task env-up               # Start PostgreSQL and MongoDB containers"
echo "  task build-host           # Build FerretDB binary"
echo "  task run                  # Run FerretDB"
echo "  task mongosh              # Connect with MongoDB shell"
echo
echo "For more information, see the documentation at: https://docs.ferretdb.io"
echo