FROM mongo:7.0.1

# If you encounter an "unknown instruction" error there,
# please update Docker to the latest version.

# see https://github.com/docker-library/mongo/issues/475
RUN <<EOF
set -ex

echo 'topsecret' > /etc/mongod_keyfile.txt
chmod 0400 /etc/mongod_keyfile.txt
chown 999:999 /etc/mongod_keyfile.txt
EOF
