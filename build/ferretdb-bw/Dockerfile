# Use production image and full tag close to the release.
FROM --platform=linux/amd64 ghcr.io/ferretdb/postgres-documentdb:17-0.106.0-ferretdb-2.5.0

# Use moving development image during development.
# FROM --platform=linux/amd64 ghcr.io/ferretdb/postgres-documentdb-dev:17-ferretdb

RUN --mount=type=cache,target=/var/cache/apt <<EOF

curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh | bash

curl -L https://pgp.mongodb.com/server-7.0.asc | apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/debian bookworm/mongodb-org/7.0 main" | tee /etc/apt/sources.list.d/mongodb-org-7.0.list

apt update
apt install -y git-lfs mongodb-mongosh tmux nano

cd /tmp
curl -OJL https://go.dev/dl/go1.24.6.linux-amd64.tar.gz
tar -xf go1.24.6.linux-amd64.tar.gz
rm go1.24.6.linux-amd64.tar.gz
mv go /usr/local/
ln -s /usr/local/go/bin/* /usr/local/bin/
go version

EOF

WORKDIR /ferretdb/build/ferretdb-bw
