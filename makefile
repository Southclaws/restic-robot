local: fast
	./restic-robot

fast:
	go build

build:
	docker build \
		-t ghcr.io/southclaws/restic-robot \
		.
