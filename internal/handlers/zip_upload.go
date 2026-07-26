package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tuantu/oj-web/internal/database/sqlcdb"
)

const (
	MaxZipSize   = 100 * 1024 * 1024 // 100MB
	MaxFileSize  = 20 * 1024 * 1024  // 20MB
	MaxTestcases = 5000
)

var (
	validInputExts  = map[string]bool{".in": true, ".inp": true, ".input": true}
	validOutputExts = map[string]bool{".out": true, ".ans": true, ".answer": true}
	ignoreFiles     = map[string]bool{"README.md": true, "Thumbs.db": true, ".DS_Store": true}
)

// Natural sort logic
func naturalSortKeys(keys []string) {
	re := regexp.MustCompile(`[0-9]+|[^0-9]+`)
	sort.Slice(keys, func(i, j int) bool {
		partsI := re.FindAllString(keys[i], -1)
		partsJ := re.FindAllString(keys[j], -1)

		for k := 0; k < len(partsI) && k < len(partsJ); k++ {
			if partsI[k] != partsJ[k] {
				numI, errI := strconv.Atoi(partsI[k])
				numJ, errJ := strconv.Atoi(partsJ[k])
				if errI == nil && errJ == nil {
					return numI < numJ
				}
				return partsI[k] < partsJ[k]
			}
		}
		return len(partsI) < len(partsJ)
	})
}

func (env *Env) HandleZipUpload(ctx context.Context, file io.Reader, size int64, problemID int64) (string, error) {
	if size > MaxZipSize {
		return "", fmt.Errorf("ZIP exceeds maximum size")
	}

	// Read zip into memory
	zipData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read zip file: %v", err)
	}

	zipReader, err := zip.NewReader(bytes.NewReader(zipData), size)
	if err != nil {
		return "", fmt.Errorf("invalid zip format: %v", err)
	}

	type TestCaseData struct {
		Input  []byte
		Output []byte
	}
	testCasesMap := make(map[string]*TestCaseData)

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		baseName := filepath.Base(f.Name)
		if ignoreFiles[baseName] || strings.HasPrefix(f.Name, "__MACOSX/") {
			continue
		}

		ext := strings.ToLower(filepath.Ext(baseName))
		// Use full path minus extension as the key to pair inputs and outputs in the same directory
		nameWithoutExt := strings.TrimSuffix(f.Name, filepath.Ext(f.Name))

		isInput := validInputExts[ext]
		isOutput := validOutputExts[ext]

		if !isInput && !isOutput {
			continue // ignore unsupported extensions
		}

		if f.UncompressedSize64 > MaxFileSize {
			return "", fmt.Errorf("file %s exceeds maximum size of 20MB", baseName)
		}

		rc, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file in zip %s: %v", baseName, err)
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read file in zip %s: %v", baseName, err)
		}

		if _, exists := testCasesMap[nameWithoutExt]; !exists {
			testCasesMap[nameWithoutExt] = &TestCaseData{}
		}

		if isInput {
			if testCasesMap[nameWithoutExt].Input != nil {
				return "", fmt.Errorf("duplicate testcase input for name: %s", nameWithoutExt)
			}
			testCasesMap[nameWithoutExt].Input = content
		} else if isOutput {
			if testCasesMap[nameWithoutExt].Output != nil {
				return "", fmt.Errorf("duplicate testcase output for name: %s", nameWithoutExt)
			}
			testCasesMap[nameWithoutExt].Output = content
		}
	}

	if len(testCasesMap) == 0 {
		return "", fmt.Errorf("ZIP does not contain any valid testcase")
	}

	if len(testCasesMap) > MaxTestcases {
		return "", fmt.Errorf("too many testcase files")
	}

	var missingInput, missingOutput []string
	var keys []string
	for k, v := range testCasesMap {
		if v.Input == nil {
			missingInput = append(missingInput, k)
		} else if v.Output == nil {
			missingOutput = append(missingOutput, k)
		} else {
			keys = append(keys, k)
		}
	}

	if len(missingInput) > 0 {
		return "", fmt.Errorf("Missing input for: %s", strings.Join(missingInput, ", "))
	}
	if len(missingOutput) > 0 {
		return "", fmt.Errorf("Missing output for: %s", strings.Join(missingOutput, ", "))
	}

	naturalSortKeys(keys)

	// Start transaction
	tx, err := env.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	qtx := env.Queries.WithTx(tx)

	err = qtx.DeleteTestCasesByProblem(ctx, problemID)
	if err != nil {
		return "", fmt.Errorf("failed to delete old test cases: %v", err)
	}

	sampleCount := 0
	hiddenCount := 0

	for i, k := range keys {
		tc := testCasesMap[k]
		isSample := strings.Contains(strings.ToLower(k), "sample")
		if isSample {
			sampleCount++
		} else {
			hiddenCount++
		}

		// Save to DB
		_, err := qtx.CreateTestCase(ctx, sqlcdb.CreateTestCaseParams{
			ProblemID:      problemID,
			OrderIndex:     int32(i + 1),
			Input:          string(tc.Input),
			ExpectedOutput: string(tc.Output),
			IsSample:       isSample,
		})
		if err != nil {
			return "", fmt.Errorf("failed to insert testcase %s: %v", k, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	return fmt.Sprintf("Import Successful! Total Testcases: %d (Sample: %d, Hidden: %d)", len(keys), sampleCount, hiddenCount), nil
}
