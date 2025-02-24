all: ferretdb_loader ferretdb

ferretdb_loader:
	$(MAKE) -C $@

ferretdb:
	$(MAKE) -C $@

clean:
	$(MAKE) -C ferretdb_loader clean
	$(MAKE) -C ferretdb clean

install:
	$(MAKE) -C ferretdb_loader install
	$(MAKE) -C ferretdb install

docker:
	docker buildx build --builder=ferretdb --tag=ferretdb-bw --load .
	docker run --platform=linux/amd64 --name=ferretdb-bw --rm -it \
		-e POSTGRES_USER=username -e POSTGRES_PASSWORD=password -e POSTGRES_DB=postgres \
		-v $(shell pwd)/../..:/ferretdb ferretdb-bw /bin/bash

docker-exec:
	docker exec -it ferretdb-bw /bin/bash

.PHONY: all ferretdb_loader ferretdb
