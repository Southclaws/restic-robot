package main

import (
	"net/http"
	"os"
	"runtime/debug"
	"time"

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

var (
	// bucketTime is the bucket configuration for time-based histograms (in ms)
	bucketTime = []float64{
		float64((5 * time.Minute).Milliseconds()),
		float64((15 * time.Minute).Milliseconds()),
		float64((30 * time.Minute).Milliseconds()),
		float64((1 * time.Hour).Milliseconds()),
		float64((2 * time.Hour).Milliseconds()),
		float64((4 * time.Hour).Milliseconds()),
		float64((8 * time.Hour).Milliseconds()),
		float64((12 * time.Hour).Milliseconds()),
	}
	// bucketFileCount is the bucket configuration for file count
	bucketFileCount = []float64{5, 50, 500, 5000, 50000, 500000, 5000000}
	// bucketFileSize is the bucket configuration for file size
	// ranging from 1 MB to 10 TB
	bucketFileSize = []float64{1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13}
)

// metrics is used to hold all the Prometheus metrics used
type metrics struct {
	backupDuration             prometheus.Histogram
	backupInfo                 prometheus.Gauge
	backupStatus               prometheus.Gauge
	backupsFailed              prometheus.Counter
	backupsSuccessful          prometheus.Counter
	backupsSuccessfulTimestamp prometheus.Gauge
	backupsTotal               prometheus.Counter
	bytesAdded                 prometheus.Histogram
	bytesProcessed             prometheus.Histogram
	filesChanged               prometheus.Histogram
	filesNew                   prometheus.Histogram
	filesProcessed             prometheus.Histogram
	filesUnmodified            prometheus.Histogram
}

// initializeMetrics configures and registers the Prometheus metrics
func (b *backup) initializeMetrics() {
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
		Buckets:   append([]float64{}, bucketTime...),
	})
	b.filesNew = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_new",
		Help:      "Amount of new files.",
		Buckets:   append([]float64{}, bucketFileCount...),
	})
	b.filesChanged = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_changed",
		Help:      "Amount of files with changes.",
		Buckets:   append([]float64{}, bucketFileCount...),
	})
	b.filesUnmodified = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_unmodified",
		Help:      "Amount of files unmodified since last backup.",
		Buckets:   append([]float64{}, bucketFileCount...),
	})
	b.filesProcessed = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_files_processed",
		Help:      "Total number of files scanned by the backup for changes.",
		Buckets:   append([]float64{}, bucketFileCount...),
	})
	b.bytesAdded = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_added_bytes",
		Help:      "Total number of bytes added to the repository.",
		Buckets:   append([]float64{}, bucketFileSize...),
	})
	b.bytesProcessed = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "backup",
		Name:      "backup_processed_bytes",
		Help:      "Total number of bytes scanned by the backup for changes",
		Buckets:   append([]float64{}, bucketFileSize...),
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
	b.backupInfo = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "backup",
		Name:        "backup_info",
		Help:        "Information about the backup process",
		ConstLabels: prometheus.Labels(getVersionInfo()),
	})
	b.backupInfo.Set(1)
	prometheus.MustRegister(
		b.backupDuration,
		b.backupInfo,
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
}

// getBuildInfo returns a value from Go's build-in debug information
func getBuildInfo(key string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == key {
				return setting.Value
			}
		}
	}
	// not found
	return ""
}

// getVersionInfo returns a list of key value pairs to become part of the info metric
func getVersionInfo() map[string]string {
	res := make(map[string]string)

	// add the hostname
	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	res["hostname"] = host

	// add the short git hash of the binary's latest commit
	revision := getBuildInfo("vcs.revision")
	if revision != "" {
		res["commit"] = revision[:7]
	}

	return res
}

func (b *backup) startMetricsServer() {
	http.Handle(b.PrometheusEndpoint, promhttp.Handler())
	logger.Info("metrics server listening at " + b.PrometheusAddress)
	err := http.ListenAndServe(b.PrometheusAddress, nil)
	logger.Fatal("metrics server closed", zap.Error(err))
}
