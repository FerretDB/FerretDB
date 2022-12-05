export CAROOT=$(PWD)
export TRUST_STORES=system

gen:
	mkcert -install
	# TODO add more hosts as needed
	mkcert -cert-file server.crt -key-file server.key localhost 127.0.0.1 ::1
	mkcert -uninstall
