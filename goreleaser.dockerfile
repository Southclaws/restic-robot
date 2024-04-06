## Dedicated Dockerfile for goreleaser, the goreleaser built binary are copied to the restic image
# https://goreleaser.com/errors/docker-build/#docker-build-failures
ARG RESTIC_VERSION
FROM restic/restic:${RESTIC_VERSION:?} AS runner
COPY restic-robot /usr/bin/restic-robot
ENTRYPOINT ["/usr/bin/restic-robot"]
