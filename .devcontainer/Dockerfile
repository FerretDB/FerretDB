# syntax=docker/dockerfile:1

FROM golang:1.24.3

# Install additional development tools
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    git git-lfs curl wget bash-completion vim jq \
    && apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Create non-root user and set up home directory
RUN groupadd --gid 1000 vscode \
    && useradd --uid 1000 --gid vscode --shell /bin/bash --create-home vscode

# Install Go tools
RUN go install -v golang.org/x/tools/gopls@latest \
    && go install -v github.com/go-delve/delve/cmd/dlv@latest \
    && go install -v honnef.co/go/tools/cmd/staticcheck@latest \
    && go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install -v github.com/cweill/gotests/gotests@latest \
    && go install -v github.com/fatih/gomodifytags@latest

# Create directories for Go cache
RUN mkdir -p /go/cache/gocache /go/cache/gomodcache \
    && chown -R vscode:vscode /go

# Create workdir and set permissions
WORKDIR /workdir
RUN chown -R vscode:vscode /workdir

# Set up environment
ENV CGO_ENABLED=1
ENV PATH="/workdir/bin:${PATH}"

# Set up Bash completion
RUN echo "source /etc/bash_completion" >> /home/vscode/.bashrc

# Add a startup script to check and install task
COPY --chown=vscode:vscode .devcontainer/setup.sh /home/vscode/setup.sh
RUN chmod +x /home/vscode/setup.sh \
    && echo "/home/vscode/setup.sh" >> /home/vscode/.bashrc

USER vscode