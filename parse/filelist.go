package main

// Handles generation of data file list.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TODO(amit): Consider extracting this to a separate package.

// dirFiles reads all the files in a given directory, recursively, and returns their paths.
// Omits directory paths.
func dirFiles(path string) ([]string, error) {
	var result []string
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			result = append(result, filepath.Join(walkPath))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// organizeInputFiles reads all files from the input diretories, creates timestamps for them, and
// sorts them according to their timestamps. Files with no timestamps are logged and omitted.
func organizeInputFiles() ([]*fileAndTime, error) {
	// Extract all file paths.
	paths := map[string]struct{}{} // Using a map to remove duplicates.
	for _, f := range args.Files {
		dir, err := dirFiles(f)
		if err != nil {
			return nil, err
		}
		for _, p := range dir {
			// TODO(amit): Change ".items" to a constant.
			if strings.HasSuffix(p, ".items") { // Ignore parsed intermediates.
				continue
			}
			paths[p] = struct{}{}
		}
	}

	// Create timestamps.
	var result []*fileAndTime
	for p := range paths {
		ts := fileTimestamp(p)
		if ts == -1 {
			pe("Skipping file with no timestamp:", p)
			continue
		}
		result = append(result, &fileAndTime{p, ts})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].time != result[j].time {
			return result[i].time < result[j].time
		}
		return result[i].file < result[j].file
	})

	return result, nil
}

// fileTimestamp infers the timestamp of a file according to its name. Returns -1 if failed.
func fileTimestamp(file string) int64 {
	match := regexp.MustCompile("(\\D|^)(20\\d{10})(\\D|$)").FindStringSubmatch(filepath.Base(file))
	if match == nil || len(match[2]) != 12 {
		return -1
	}
	digits := match[2]
	year, _ := strconv.ParseInt(digits[0:4], 10, 64)
	month, _ := strconv.ParseInt(digits[4:6], 10, 64)
	day, _ := strconv.ParseInt(digits[6:8], 10, 64)
	hour, _ := strconv.ParseInt(digits[8:10], 10, 64)
	minute, _ := strconv.ParseInt(digits[10:12], 10, 64)
	t := time.Date(int(year), time.Month(month), int(day), int(hour),
		int(minute), 0, 0, time.UTC)

	return t.Unix()
}

// Represents a file and its timestamp.
type fileAndTime struct {
	file string
	time int64
}

// For debugging.
func (f *fileAndTime) String() string {
	return fmt.Sprintf("{%v %v}", f.file, f.time)
}
