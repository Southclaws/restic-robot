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
- `RUN_ON_BOOT`: run a backup on startup
- `PROMETHEUS_ENDPOINT`: metrics endpoint
- `PROMETHEUS_ADDRESS`: metrics host:port

Prometheus metrics:

- `backups_all_total`: The total number of backups attempted, including failures.
- `backups_successful_total`: The total number of backups that succeeded.
- `backups_failed_total`: The total number of backups that failed.
- `backup_duration_milliseconds`: The duration of backups in milliseconds.
- `backup_files_new`: Amount of new files.
- `backup_files_changed`: Amount of files with changes.
- `backup_files_unmodified`: Amount of files unmodified since last backup.
- `backup_files_processed`: Total number of files scanned by the backup for changes.
- `backup_added_bytes`: Total number of bytes added to the repository.
- `backup_processed_bytes`: Total number of bytes scanned by the backup for changes

It's that simple!

## Docker Compose

Stick this in with your other compose services for instant backups!

```yml
services:
  #
  # your stuff etc...
  #

  backup:
    image: southclaws/restic-robot
    restart: always
    environment:
      # every day at 2am
      SCHEDULE: 0 0 2 * * *
      RESTIC_REPOSITORY: my_service_repository
      RESTIC_PASSWORD: ${MY_SERVICE_RESTIC_PASSWORD}
      # restic-robot runs `restic backup ${RESTIC_ARGS}`
      # so this is where you specify the directory and any other args.
      RESTIC_ARGS: /data
      B2_ACCOUNT_ID: ${B2_ACCOUNT_ID}
      B2_ACCOUNT_KEY: ${B2_ACCOUNT_KEY}
    volumes:
      # Bind whatever directories to the backup container.
      # You can safely bind the same directory to multiple containers.
      - "/container_data/blog/wordpress:/data/wordpress"
```
