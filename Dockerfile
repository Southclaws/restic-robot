ARG RESTIC_VERSION=latest

FROM golang:alpine AS builder

WORKDIR /restic-robot
ADD . .
ENV GO111MODULE=on
RUN apk add git
RUN apk add gcc
RUN apk add musl-dev
RUN go mod tidy
RUN go build

FROM restic/restic:${RESTIC_VERSION:?} AS runner

COPY --from=builder /restic-robot/restic-robot /usr/bin/restic-robot

ENTRYPOINT ["restic-robot"]
