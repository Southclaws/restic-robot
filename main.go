package main

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

type backup struct {
	Schedule           string `required:"true"    envconfig:"SCHEDULE"`            // cron schedule
	Repository         string `required:"true"    envconfig:"RESTIC_REPOSITORY"`   // repository name
	Password           string `required:"true"    envconfig:"RESTIC_PASSWORD"`     // repository password
	Args               string `                   envconfig:"RESTIC_ARGS"`         // additional args for backup command
	RunOnBoot          bool   `                   envconfig:"RUN_ON_BOOT"`         // run a backup on startup
	PrometheusEndpoint string `default:"/metrics" envconfig:"PROMETHEUS_ENDPOINT"` // metrics endpoint
	PrometheusAddress  string `default:":8080"    envconfig:"PROMETHEUS_ADDRESS"`  // metrics host:port

	backupsTotal      prometheus.Counter
	backupsSuccessful prometheus.Counter
	backupsFailed     prometheus.Counter
	backupDuration    prometheus.Histogram
	filesNew          prometheus.Histogram
	filesChanged      prometheus.Histogram
	filesUnmodified   prometheus.Histogram
	filesProcessed    prometheus.Histogram
	bytesAdded        prometheus.Histogram
	bytesProcessed    prometheus.Histogram
}

var (
	matchExists     = regexp.MustCompile(`.*already (exists|initialized).*`)
	matchFileStats  = regexp.MustCompile(`Files:\s*([0-9.]*) new,\s*([0-9.]*) changed,\s*([0-9.]*) unmodified`)
	matchAddedBytes = regexp.MustCompile(`Added to the repo: ([0-9.]+) (\w+)`)
	matchProcessed  = regexp.MustCompile(`processed ([0-9.]*) files, ([0-9.]+) (\w+)`)
)

type stats struct {
	filesNew        int
	filesChanged    int
	filesUnmodified int
	filesProcessed  int
	bytesAdded      int
	bytesProcessed  int
}

func main() {
	b := backup{}
	err := envconfig.Process("", &b)
	if err != nil {
		logger.Fatal("failed to configure", zap.Error(err))
	}

	err = b.Ensure()
	if err != nil {
		logger.Fatal("failed to ensure repository", zap.Error(err))
	}

	go b.startMetricsServer()

	cr := cron.New()
	err = cr.AddJob(b.Schedule, &b)
	if err != nil {
		logger.Fatal("failed to schedule task", zap.Error(err))
	}
	if b.RunOnBoot {
		b.Run()
	}
	cr.Run()
}

// Run performs the backup
func (b *backup) Run() {
	logger.Info("backup started")
	startTime := time.Now()

	cmd := exec.Command("restic", "backup", "-q", "-v", b.Args)
	errors := bytes.NewBuffer(nil)
	output := bytes.NewBuffer(nil)
	cmd.Stderr = errors
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		logger.Error("failed to run backup",
			zap.Error(err),
			zap.String("output", errors.String()))
		b.backupsFailed.Inc()
	} else {
		d := time.Since(startTime)

		statistics, err := extractStats(output.String())
		if err != nil {
			logger.Warn("failed to extract statistics from command output",
				zap.Error(err))
		}

		logger.Info("backup completed",
			zap.Duration("duration", d),
			zap.Int("filesNew", statistics.filesNew),
			zap.Int("filesChanged", statistics.filesChanged),
			zap.Int("filesUnmodified", statistics.filesUnmodified),
			zap.Int("filesProcessed", statistics.filesProcessed),
			zap.Int("bytesAdded", statistics.bytesAdded),
			zap.Int("bytesProcessed", statistics.bytesProcessed),
		)

		b.backupsSuccessful.Inc()
		b.backupDuration.Observe(float64(d.Nanoseconds() * 1000))
		b.filesNew.Observe(float64(statistics.filesNew))
		b.filesChanged.Observe(float64(statistics.filesChanged))
		b.filesUnmodified.Observe(float64(statistics.filesUnmodified))
		b.filesProcessed.Observe(float64(statistics.filesProcessed))
		b.bytesAdded.Observe(float64(statistics.bytesAdded))
		b.bytesProcessed.Observe(float64(statistics.bytesProcessed))
	}

	b.backupsTotal.Inc()
	return
}

func extractStats(s string) (result stats, err error) {
	fileStats := matchFileStats.FindAllStringSubmatch(s, -1)
	if len(fileStats[0]) != 4 {
		err = errors.Errorf("matchFileStats expected 4, got %d", len(fileStats[0]))
		return
	}
	result.filesNew, _ = strconv.Atoi(fileStats[0][1])        //nolint:errcheck
	result.filesChanged, _ = strconv.Atoi(fileStats[0][2])    //nolint:errcheck
	result.filesUnmodified, _ = strconv.Atoi(fileStats[0][3]) //nolint:errcheck

	addedBytes := matchAddedBytes.FindAllStringSubmatch(s, -1)
	if len(addedBytes[0]) != 3 {
		err = errors.Errorf("matchAddedBytes expected 3, got %d", len(addedBytes[0]))
		return
	}
	amount, _ := strconv.ParseFloat(addedBytes[0][1], 64) //nolint:errcheck
	// restic doesn't use a comma to denote thousands
	amount *= 1000
	result.bytesAdded = convert(int(amount), addedBytes[0][2])

	filesProcessed := matchProcessed.FindAllStringSubmatch(s, -1)
	if len(filesProcessed[0]) != 4 {
		err = errors.Errorf("filesProcessed expected 4, got %d", len(filesProcessed[0]))
		return
	}
	result.filesProcessed, _ = strconv.Atoi(filesProcessed[0][1]) //nolint:errcheck
	amount, _ = strconv.ParseFloat(filesProcessed[0][2], 64)      //nolint:errcheck
	amount *= 1000
	result.bytesProcessed = convert(int(amount), filesProcessed[0][3])

	return
}

func convert(b int, unit string) (result int) {
	switch unit {
	case "TiB":
		result = b * (1 << 40)
	case "GiB":
		result = b * (1 << 30)
	case "MiB":
		result = b * (1 << 20)
	case "KiB":
		result = b * (1 << 10)
	}
	return
}

// Ensure will create a repository if it does not already exist
func (b *backup) Ensure() (err error) {
	logger.Info("ensuring backup repository exists")
	cmd := exec.Command("restic", "init")
	out := bytes.NewBuffer(nil)
	cmd.Stderr = out
	err = cmd.Run()
	if err != nil {
		if matchExists.MatchString(strings.Trim(out.String(), " \n\r")) {
			logger.Info("repository exists")
			return nil
		}
		return errors.Wrap(err, out.String())
	}
	logger.Info("successfully created repository")
	return
}
