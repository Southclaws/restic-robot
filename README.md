# Restic Robot

Backups done right... by robots!

This is a small and simple wrapper application for [Restic](https://github.com/restic/restic/) that provides:

- Automatically creates repository if it doesn't already exist
- Scheduled backups - no need for system-wide cron
- Prometheus metrics- know when your backups don't run!
- JSON logs - for the robots!
- Pre/post shell script hooks for custom behaviour! (Thanks @opthomas-prime!)

## Usage

Just `go build` and run it, or, if you're into Docker, `ghcr.io/southclaws/restic-robot`.

Environment variables:

- `SCHEDULE`: cron schedule
- `RESTIC_REPOSITORY`: repository name
- `RESTIC_PASSWORD`: repository password
- `RESTIC_ARGS`: additional args for backup command
- `RUN_ON_BOOT`: run a backup on startup
- `PROMETHEUS_ENDPOINT`: metrics endpoint
- `PROMETHEUS_ADDRESS`: metrics host:port
- `PRE_COMMAND`: A shell command to run before a backup starts
- `POST_COMMAND`: A shell command to run if the backup completes successfully
- `ERROR_COMMAND`: A shell command to run if the backup errors. For example, to send a notification to a Slack channel on backup failure, you could set it to a curl command that posts to your Slack webhook.
- `TRIGGER_ENDPOINT`: manual trigger endpoint

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

### Manual backups

Sometimes backups are required out-of-band - e.g. before some manual changes to a system
are made. Instead of running `restic` manually, or to edit the cron schedule for a single
run, you can trigger a manual backup by sending an HTTP POST request to the configured
`TRIGGER_ENDPOINT` (defaulting to `http://localhost:8080/trigger`). It reuses the listen
address configured with `PROMETHEUS_ADDRESS`. If the endpoint is set to an empty string,
manual backups are disabled.

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
