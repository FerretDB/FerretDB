export CAROOT=$(PWD)
export TRUST_STORES=system

gen:
	mkcert -install
	# TODO add more hosts as needed
	mkcert -cert-file server.pem -key-file server.pem   localhost 127.0.0.1 ::1 host.docker.internal
	mkcert -uninstall
