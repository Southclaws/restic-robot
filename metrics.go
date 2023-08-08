package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

const (
	// backupStatusIdle indicates there is no backup in progress
	backupStatusIdle = 0
	// backupStatusFailed indicates no running backup and the last one failed
	backupStatusFailed = -1
	// backupStatusRunning indicates the backup is currently in progress
	backupStatusRunning = 1
)

func (b *backup) startMetricsServer() {

	b.backupsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "backup",
		Name:      "backups_all_total",
		Help:      "The total number of backups attempted, including failures.",
	})
	b.backupsSuccessful = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "backup",
		Name:      "backups_successful_total",
		Help:      "The total number of backups that succeeded.",
	})
	b.backupsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "backup",
		Name:      "backups_failed_total",
		Help:      "The total number of backups that failed.",
	})
	b.backupDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_duration_milliseconds",
		Help:      "The duration of backups in milliseconds.",
	})
	b.filesNew = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_new",
		Help:      "Amount of new files.",
	})
	b.filesChanged = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_changed",
		Help:      "Amount of files with changes.",
	})
	b.filesUnmodified = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_unmodified",
		Help:      "Amount of files unmodified since last backup.",
	})
	b.filesProcessed = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_processed",
		Help:      "Total number of files scanned by the backup for changes.",
	})
	b.bytesAdded = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_added_bytes",
		Help:      "Total number of bytes added to the repository.",
	})
	b.bytesProcessed = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_processed_bytes",
		Help:      "Total number of bytes scanned by the backup for changes",
	})
	b.backupsSuccessfulTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "backup",
		Name:      "backup_successful_timestamp",
		Help:      "Timestamp of last successful backup",
	})
	b.backupStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "backup",
		Name:      "backup_status",
		Help:      "Backup status (1 = backing up, 0 = idle, -1 = idle after failed backup)",
	})
	prometheus.MustRegister(
		b.backupDuration,
		b.backupStatus,
		b.backupsFailed,
		b.backupsSuccessful,
		b.backupsSuccessfulTimestamp,
		b.backupsTotal,
		b.bytesAdded,
		b.bytesProcessed,
		b.filesChanged,
		b.filesNew,
		b.filesProcessed,
		b.filesUnmodified,
	)

	http.Handle(b.PrometheusEndpoint, promhttp.Handler())
	err := http.ListenAndServe(b.PrometheusAddress, nil)
	logger.Fatal("metrics server closed", zap.Error(err))
}
