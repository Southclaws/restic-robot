include .env
export

local: fast
	./restic-robot

fast:
	go build

build:
	docker build \
		--build-arg RESTIC_VERSION=$(RESTIC_VERSION) \
		-t ghcr.io/southclaws/restic-robot \
		.
