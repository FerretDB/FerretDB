# Development Container for FerretDB

This directory contains configuration files for setting up a development environment for FerretDB using Visual Studio Code's [Development Containers](https://code.visualstudio.com/docs/devcontainers/containers) or [GitHub Codespaces](https://github.com/features/codespaces).

## Features

- Preconfigured Go development environment
- PostgreSQL and MongoDB containers for backend testing
- All necessary tools pre-installed
- VS Code extensions for Go development, linting, etc.
- Shared volume for Go module cache to improve build speed

## Usage

### With VS Code

1. Install the [Remote - Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) extension
2. Open this repository in VS Code
3. When prompted "Reopen in Container", click it (or use the Command Palette: `F1` -> "Remote-Containers: Reopen in Container")
4. Wait for the container to build (this might take a few minutes the first time)

### With GitHub Codespaces

1. Go to the GitHub repository
2. Click the "Code" button
3. Select the "Codespaces" tab
4. Click "Create codespace on main"

## Development Workflow

Once inside the container, you can use the following commands:

- `task -l` - List all available tasks
- `task build-host` - Build FerretDB
- `task run` - Run FerretDB
- `task mongosh` - Connect to FerretDB with MongoDB shell

The development workflow inside the container is the same as the workflow described in the main [CONTRIBUTING.md](../CONTRIBUTING.md) file, except that everything runs inside the container.

## Configuration

The container configuration is defined in:

- `devcontainer.json` - VS Code and Codespaces configuration
- `Dockerfile` - Container definition
- `../docker-compose.devcontainer.yml` - Docker Compose configuration for the devcontainer

## Troubleshooting

- **Container fails to start**: Ensure Docker is running and you have enough resources allocated
- **Cannot connect to PostgreSQL**: The devcontainer uses the network of the postgres service, so you should be able to connect to it at host `postgres`
- **Volume performance issues**: The container uses volume mounting to preserve your workspace and Go cache. On some systems, this may be slow. Consider using a `delegated` or `cached` volume mount in the Docker Compose file.