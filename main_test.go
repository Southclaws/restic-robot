package main

import (
	"bytes"
	"fmt"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

func Test_extractJsonStats(t *testing.T) {
	tests := []struct {
		input      string
		wantResult stats
	}{
		{`{"message_type":"summary","files_new":56,"files_changed":2,"files_unmodified":2,"dirs_new":0,"dirs_changed":0,"dirs_unmodified":0,"data_blobs":35,"tree_blobs":1,"data_added":169009,"total_files_processed":58,"total_bytes_processed":102624133120}
`, stats{
			filesNew:        56,
			filesChanged:    2,
			filesUnmodified: 2,
			filesProcessed:  58,
			bytesAdded:      169009,
			bytesProcessed:  102624133120,
		}},
	}
	for ii, tt := range tests {
		t.Run(fmt.Sprint(ii), func(t *testing.T) {
			inputBuffer := bytes.NewBufferString(tt.input)
			gotResult, err := extractJsonStats(inputBuffer)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
