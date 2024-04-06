## Dedicated Dockerfile for goreleaser, the goreleaser built binary are copied to the restic image
# https://goreleaser.com/errors/docker-build/#docker-build-failures
FROM restic/restic
COPY restic-robot /usr/bin/restic-robot
ENTRYPOINT ["/usr/bin/restic-robot"]
