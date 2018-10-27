local: fast
	./restic-robot

fast:
	go build

build:
	docker build \
		-t southclaws/restic-robot \
		.
