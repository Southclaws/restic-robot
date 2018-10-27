# Restic Robot

Backups done right... by robots!

This is a small and simple wrapper application for [Restic](https://github.com/restic/restic/) that provides:

- Automatically creates repository if it doesn't already exist
- Scheduled backups - no need for system-wide cron
- Prometheus metrics- know when your backups don't run!
- JSON logs - for the robots!

## Usage

Just `go build` and run it, or, if you're into Docker, `southclaws/restic-robot`.

Environment variables:

- `SCHEDULE`: cron schedule
- `RESTIC_REPOSITORY`: repository name
- `RESTIC_PASSWORD`: repository password
- `RESTIC_ARGS`: additional args for backup command
- `PROMETHEUS_ENDPOINT`: metrics endpoint
- `PROMETHEUS_ADDRESS`: metrics host:port

It's that simple!
