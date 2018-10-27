FROM golang:alpine AS builder

WORKDIR /restic-robot
ADD . .
ENV GO111MODULE=on
RUN apk add git
RUN go mod tidy
RUN go build

FROM restic/restic AS runner

COPY --from=builder /restic-robot/restic-robot /usr/bin/restic-robot

ENTRYPOINT ["restic-robot"]
