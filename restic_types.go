package main

// Restic Types from JSON output: https://restic.readthedocs.io/en/stable/075_scripting.html#json-output

// BackupMessage represents a general message from Restic backup, with a MessageType for type assertion.
type BackupMessage struct {
	MessageType string `json:"message_type"`
}

// BackupStatusMessage represents the "status" messages emitted during backup operations.
type BackupStatusMessage struct {
	MessageType      string   `json:"message_type"`
	SecondsElapsed   float64  `json:"seconds_elapsed"`
	SecondsRemaining float64  `json:"seconds_remaining,omitempty"`
	PercentDone      float64  `json:"percent_done"`
	TotalFiles       int      `json:"total_files"`
	FilesDone        int      `json:"files_done"`
	TotalBytes       int64    `json:"total_bytes"`
	BytesDone        int64    `json:"bytes_done"`
	ErrorCount       int      `json:"error_count"`
	CurrentFiles     []string `json:"current_files,omitempty"`
}

// BackupErrorMessage represents the "error" messages that provide details on errors encountered.
type BackupErrorMessage struct {
	MessageType string `json:"message_type"`
	Error       string `json:"error"`
	During      string `json:"during"`
	Item        string `json:"item"`
}

// BackupVerboseStatusMessage provides detailed progress updates, including specifics about files being backed up.
type BackupVerboseStatusMessage struct {
	MessageType  string  `json:"message_type"`
	Action       string  `json:"action"`
	Item         string  `json:"item"`
	Duration     float64 `json:"duration"`
	DataSize     int64   `json:"data_size"`
	MetadataSize int64   `json:"metadata_size"`
	TotalFiles   int     `json:"total_files"`
}

// BackupSummaryMessage represents the summary of a completed backup operation.
type BackupSummaryMessage struct {
	MessageType         string  `json:"message_type"`
	FilesNew            int     `json:"files_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           int64   `json:"data_added"`
	TotalFilesProcessed int     `json:"total_files_processed"`
	TotalBytesProcessed int64   `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"`
	SnapshotID          string  `json:"snapshot_id"`
}
