package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"go.uber.org/zap"
)

type backup struct {
	Schedule           string `required:"true"    envconfig:"SCHEDULE"`            // cron schedule
	Repository         string `required:"true"    envconfig:"RESTIC_REPOSITORY"`   // repository name
	Password           string `required:"true"    envconfig:"RESTIC_PASSWORD"`     // repository password
	Args               string `                   envconfig:"RESTIC_ARGS"`         // additional args for backup command
	RunOnBoot          bool   `                   envconfig:"RUN_ON_BOOT"`         // run a backup on startup
	TriggerEndpoint    string `default:"/trigger" envconfig:"TRIGGER_ENDPOINT"`    // trigger endpoint
	PrometheusEndpoint string `default:"/metrics" envconfig:"PROMETHEUS_ENDPOINT"` // metrics endpoint
	PrometheusAddress  string `default:":8080"    envconfig:"PROMETHEUS_ADDRESS"`  // metrics host:port
	PreCommand         string `                   envconfig:"PRE_COMMAND"`         // command to execute before restic is executed
	PostCommand        string `                   envconfig:"POST_COMMAND"`        // command to execute after restic was executed (successfully)

	// lock is used to prevent concurrent backups from happening
	lock sync.Mutex
	// metrics defines all the different Prometheus metrics in use
	metrics
}

var (
	matchExists = regexp.MustCompile(`.*already (exists|initialized).*`)
)

type stats struct {
	filesNew        int
	filesChanged    int
	filesUnmodified int
	filesProcessed  int
	bytesAdded      int64
	bytesProcessed  int64
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
	b.initializeMetrics()
	if b.PrometheusAddress != "" {
		b.setupTrigger()
		go b.startMetricsServer()
	} else {
		logger.Info("metrics and manual trigger disabled")
	}

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
	// prevent concurrent backups from happening
	if !b.lock.TryLock() {
		logger.Warn("backup is already running")
		return
	}
	// ensure lock is released after backup
	defer b.lock.Unlock()

	logger.Info("backup started")
	startTime := time.Now()
	// hold the backup success
	success := false
	b.backupStatus.Set(backupStatusRunning)
	// process metrics after backup completed
	defer func() {
		if success {
			// last backup succeeded
			b.backupsSuccessful.Inc()
			b.backupStatus.Set(backupStatusIdle)
			b.backupsSuccessfulTimestamp.SetToCurrentTime()
		} else {
			// last backup failed
			b.backupsFailed.Inc()
			b.backupStatus.Set(backupStatusFailed)
		}
		b.backupsTotal.Inc()
	}()

	// execute pre-command (if configured)
	if len(b.PreCommand) > 0 {
		if stdout, err := b.executePreCommand(); err != nil {
			logger.Error("failed to execute pre-command: " + err.Error())
			return
		} else {
			logger.Info("output of pre-command: " + *stdout)
		}
	}

	// execute restic backup
	cmd := exec.Command("restic", append([]string{"backup", "--json"}, parseArg(b.Args)...)...)
	errbuf := bytes.NewBuffer(nil)
	outbuf := bytes.NewBuffer(nil)
	cmd.Stderr = errbuf
	cmd.Stdout = outbuf

	if err := cmd.Run(); err != nil {
		logger.Error("failed to run backup",
			zap.Error(err),
			zap.String("output", errbuf.String()))
		b.backupStatus.Set(backupStatusFailed)
		b.backupsFailed.Inc()
		b.backupsTotal.Inc()
		return
	}

	if len(b.PostCommand) > 0 {
		if stdout, err := b.executePostCommand(); err != nil {
			logger.Error("failed to execute post-command: " + err.Error())
			return
		} else {
			logger.Info("output of post-command: " + *stdout)
		}
	}

	d := time.Since(startTime)

	statistics, err := extractJsonStats(outbuf)
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
		zap.Int64("bytesAdded", statistics.bytesAdded),
		zap.Int64("bytesProcessed", statistics.bytesProcessed),
	)

	// indicate backup success
	success = true

	// process result and update metrics
	b.backupDuration.Observe(float64(d.Milliseconds()))
	b.filesNew.Observe(float64(statistics.filesNew))
	b.filesChanged.Observe(float64(statistics.filesChanged))
	b.filesUnmodified.Observe(float64(statistics.filesUnmodified))
	b.filesProcessed.Observe(float64(statistics.filesProcessed))
	b.bytesAdded.Observe(float64(statistics.bytesAdded))
	b.bytesProcessed.Observe(float64(statistics.bytesProcessed))
}

func extractJsonStats(outbuf *bytes.Buffer) (result stats, err error) {
	reader := bufio.NewReader(outbuf)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("error reading output buffer", zap.Error(err))
			return result, err
		}

		var msg BackupMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			logger.Error("error unmarshalling JSON", zap.ByteString("line", line), zap.Error(err))
			return result, err
		}

		switch msg.MessageType {
		case "summary":
			var summary BackupSummaryMessage
			if err := json.Unmarshal(line, &summary); err != nil {
				logger.Error("error unmarshalling summary message", zap.ByteString("line", line), zap.Error(err))
				return result, err
			}
			result.filesNew = summary.FilesNew
			result.filesChanged = summary.FilesChanged
			result.filesUnmodified = summary.FilesUnmodified
			result.filesProcessed = summary.TotalFilesProcessed
			result.bytesAdded = summary.DataAdded
			result.bytesProcessed = summary.TotalBytesProcessed
		case "status":
			logger.Debug("received status update", zap.ByteString("line", line))
		case "error":
			var errorMsg BackupErrorMessage
			if err := json.Unmarshal(line, &errorMsg); err != nil {
				logger.Error("error unmarshalling error message", zap.ByteString("line", line), zap.Error(err))
			} else {
				logger.Error("backup error", zap.String("error", errorMsg.Error), zap.String("during", errorMsg.During), zap.String("item", errorMsg.Item))
			}
		}
	}
	return result, nil
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
