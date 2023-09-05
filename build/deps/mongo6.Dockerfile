FROM mongo:6.0.9

# see https://github.com/docker-library/mongo/issues/475
RUN <<EOF
set -ex

echo 'topsecret' > /etc/mongod_keyfile.txt
chmod 0400 /etc/mongod_keyfile.txt
chown 999:999 /etc/mongod_keyfile.txt
EOF
