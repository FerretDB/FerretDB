export CAROOT=$(PWD)
export TRUST_STORES=system

gen:
	mkcert -install
	mkcert -cert-file server-cert.pem -key-file server-key.pem localhost 127.0.0.1 ::1 host.docker.internal
	cat server-cert.pem server-key.pem > server.pem
	mkcert -uninstall
