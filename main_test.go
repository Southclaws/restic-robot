package main

import (
	"fmt"
	"testing"

	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
)

func Test_extractStats(t *testing.T) {
	tests := []struct {
		input      string
		wantResult stats
	}{
		{`open repository
lock repository
load index files
using parent snapshot 38401629
start scan on [./test/]
start backup on [./test/]
scan finished in 0.797s: 58 files, 97.870 MiB

Files:          56 new,     2 changed,     2 unmodified
Dirs:            0 new,     0 changed,     0 unmodified
Data Blobs:     35 new
Tree Blobs:      1 new
Added to the repo: 169.009 KiB

processed 58 files, 97.870 MiB in 0:00
snapshot c7693989 saved`, stats{
			filesNew:        56,
			filesChanged:    2,
			filesUnmodified: 2,
			filesProcessed:  58,
			bytesAdded:      173065216,
			bytesProcessed:  102624133120,
		}},
	}
	for ii, tt := range tests {
		t.Run(fmt.Sprint(ii), func(t *testing.T) {
			gotResult, err := extractStats(tt.input)
			if err != nil {
				t.Error(err)
			}
			assert.Equal(t, tt.wantResult, gotResult)
		})
	}
}
