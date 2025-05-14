package collector

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"projectson/config"
	"projectson/utils"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// FileCollector handles the logic of collecting and processing files.
type FileCollector struct {
	Config         *config.Config
	excludeRegexps map[string]*regexp.Regexp
}

// OutputJSON represents the structure of the final JSON output.
type OutputJSON struct {
	ProjectFiles []ProcessedFile `json:"project_files"`
}

// NewFileCollector creates a new FileCollector instance.
func NewFileCollector(cfg *config.Config) (*FileCollector, error) {
	fc := &FileCollector{
		Config:         cfg,
		excludeRegexps: make(map[string]*regexp.Regexp),
	}
	for _, pattern := range cfg.ExcludePatterns {
		if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
			regexStr := pattern[1 : len(pattern)-1]
			compiledRegexp, err := regexp.Compile(regexStr)
			if err != nil {
				fmt.Printf("Warning: Invalid regex exclude pattern '%s': %v\n", regexStr, err)
			} else {
				fc.excludeRegexps[pattern] = compiledRegexp
			}
		}
	}
	return fc, nil
}

func (fc *FileCollector) parseInclude() []config.ParsedIncludeEntry {
	parsed := []config.ParsedIncludeEntry{}
	for _, entry := range fc.Config.Include {
		if strings.TrimSpace(entry) == "" {
			continue
		}
		path := entry
		mode := "both"
		isDirOnlyFiles := false
		if strings.HasSuffix(entry, "/*") {
			path = strings.TrimSuffix(entry, "/*")
			isDirOnlyFiles = true
		}
		if strings.Contains(path, ":") {
			parts := strings.SplitN(path, ":", 2)
			path = strings.TrimSpace(parts[0])
			param := strings.TrimSpace(strings.ToLower(parts[1]))
			if param == "path" || param == "content" {
				mode = param
			}
		}
		parsed = append(parsed, config.ParsedIncludeEntry{Path: path, Mode: mode, IsDirOnlyFiles: isDirOnlyFiles})
	}
	return parsed
}

func (fc *FileCollector) isExcluded(path string, isDir bool) (bool, error) {
	baseName := filepath.Base(path)
	relPath, err := filepath.Rel(fc.Config.Root, path)
	if err != nil {
		relPath = baseName
	}

	for _, pattern := range fc.Config.ExcludePatterns {
		if compiledRegexp, ok := fc.excludeRegexps[pattern]; ok && compiledRegexp != nil {
			if compiledRegexp.MatchString(baseName) || compiledRegexp.MatchString(relPath) || compiledRegexp.MatchString(path) {
				return true, nil
			}
		} else if !strings.HasPrefix(pattern, "/") {
			matchBase, _ := filepath.Match(pattern, baseName)
			if matchBase {
				return true, nil
			}
			if strings.ContainsRune(pattern, os.PathSeparator) || strings.ContainsRune(pattern, '/') {
				matchPath, _ := filepath.Match(pattern, relPath)
				if matchPath {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func (fc *FileCollector) matchFormat(filename string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	for _, format := range fc.Config.Formats {
		if ext == format {
			return true
		}
	}
	return false
}

func (fc *FileCollector) PreviewFiles() ([]FileEntry, error) {
	var entries []FileEntry
	rootBase := filepath.Base(fc.Config.Root)
	includes := fc.parseInclude()
	foundFiles := make(map[string]FileEntry)

	for _, include := range includes {
		absIncludePath := filepath.Join(fc.Config.Root, include.Path)
		info, err := os.Stat(absIncludePath)
		if os.IsNotExist(err) {
			fmt.Printf("Warning: include path not found: %s\n", absIncludePath)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("error stating include path %s: %w", absIncludePath, err)
		}

		if include.IsDirOnlyFiles && info.IsDir() {
			files, err := os.ReadDir(absIncludePath)
			if err != nil {
				return nil, fmt.Errorf("error reading directory %s: %w", absIncludePath, err)
			}
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(absIncludePath, file.Name())
					excluded, _ := fc.isExcluded(filePath, false)
					if !excluded && fc.matchFormat(file.Name()) {
						fileInfo, statErr := file.Info()
						if statErr != nil {
							fmt.Printf("Warning: could not stat file %s: %v\n", filePath, statErr)
							continue
						}
						relPath, _ := filepath.Rel(fc.Config.Root, filePath)
						entry := FileEntry{
							Path:         filepath.Join(rootBase, relPath),
							OriginalPath: relPath,
							SourcePath:   filePath,
							Mode:         include.Mode,
							Size:         fileInfo.Size(),
							Format:       strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Name()), ".")),
						}
						foundFiles[entry.Path] = entry
					}
				}
			}
		} else if info.IsDir() { // Recursive walk
			err := filepath.WalkDir(absIncludePath, func(currentPath string, d fs.DirEntry, errWalk error) error {
				if errWalk != nil {
					fmt.Printf("Warning: error accessing path %q: %v\n", currentPath, errWalk)
					return errWalk
				}
				excluded, _ := fc.isExcluded(currentPath, d.IsDir())
				if excluded {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				if !d.IsDir() && fc.matchFormat(d.Name()) {
					fileInfo, statErr := d.Info()
					if statErr != nil {
						fmt.Printf("Warning: could not stat file %s: %v\n", currentPath, statErr)
						return nil
					}
					relPath, _ := filepath.Rel(fc.Config.Root, currentPath)
					entry := FileEntry{
						Path:         filepath.Join(rootBase, relPath),
						OriginalPath: relPath,
						SourcePath:   currentPath,
						Mode:         include.Mode,
						Size:         fileInfo.Size(),
						Format:       strings.ToLower(strings.TrimPrefix(filepath.Ext(d.Name()), ".")),
					}
					foundFiles[entry.Path] = entry
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("error walking directory %s: %w", absIncludePath, err)
			}
		} else { // Single file
			excluded, _ := fc.isExcluded(absIncludePath, false)
			if !excluded && fc.matchFormat(info.Name()) {
				relPath, _ := filepath.Rel(fc.Config.Root, absIncludePath)
				entry := FileEntry{
					Path:         filepath.Join(rootBase, relPath),
					OriginalPath: relPath,
					SourcePath:   absIncludePath,
					Mode:         include.Mode,
					Size:         info.Size(),
					Format:       strings.ToLower(strings.TrimPrefix(filepath.Ext(info.Name()), ".")),
				}
				foundFiles[entry.Path] = entry
			}
		}
	}

	for _, entry := range foundFiles {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

func (fc *FileCollector) ApplyContentExclusions(content string, fileExt string) (string, error) {
	modifiedContent := content
	if len(fc.Config.ContentExclusions) == 0 {
		return content, nil
	}
	dotFileExt := "." + fileExt // e.g. ".vue"

	for _, exclusion := range fc.Config.ContentExclusions {
		applies := false
		if exclusion.FilePattern == "*" {
			applies = true
		} else {
			matchExt, _ := filepath.Match(exclusion.FilePattern, fileExt)
			matchDotExt, _ := filepath.Match(exclusion.FilePattern, dotFileExt)
			if matchExt || matchDotExt {
				applies = true
			}
		}

		if !applies {
			continue
		}

		switch exclusion.Type {
		case "delimiters":
			if exclusion.Start != "" && exclusion.End != "" {
				regexPattern := fmt.Sprintf("(?s)%s.*?%s", regexp.QuoteMeta(exclusion.Start), regexp.QuoteMeta(exclusion.End))
				re, err := regexp.Compile(regexPattern)
				if err != nil {
					fmt.Printf("Warning: Invalid delimiter regex pattern derived: %v\n", err)
					continue
				}
				modifiedContent = re.ReplaceAllString(modifiedContent, "")
			}
		case "regexp":
			if exclusion.Pattern != "" {
				pattern := exclusion.Pattern
				if !strings.HasPrefix(pattern, "(?s)") && !strings.Contains(pattern, `\A`) && !strings.Contains(pattern, `\z`) {
					if !strings.HasPrefix(pattern, "(?i)") && !strings.HasPrefix(pattern, "(?m)") {
						pattern = "(?s)" + pattern
					}
				}
				re, err := regexp.Compile(pattern)
				if err != nil {
					fmt.Printf("Warning: Invalid content exclusion regex pattern '%s': %v\n", exclusion.Pattern, err)
					continue
				}
				modifiedContent = re.ReplaceAllString(modifiedContent, "")
			}
		default:
			fmt.Printf("Warning: Unknown content exclusion type: %s\n", exclusion.Type)
		}
	}
	return modifiedContent, nil
}

func (fc *FileCollector) processFile(entry FileEntry) (ProcessedFile, error) {
	result := make(ProcessedFile)
	if entry.Mode == "path" || entry.Mode == "both" {
		result["path"] = entry.Path
	}

	if entry.Mode == "content" || entry.Mode == "both" {
		contentBytes, err := os.ReadFile(entry.SourcePath)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", entry.SourcePath, err)
		}
		content := string(contentBytes)
		content, err = fc.ApplyContentExclusions(content, entry.Format)
		if err != nil {
			return nil, fmt.Errorf("applying content exclusions to %s: %w", entry.SourcePath, err)
		}
		spaceRe := regexp.MustCompile(`\s+`)
		compressedContent := spaceRe.ReplaceAllString(content, " ")
		result["content"] = strings.TrimSpace(compressedContent)
	}

	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func (fc *FileCollector) Run(progressCallback func(current, total int)) (int, string, error) {
	filesToProcess, err := fc.PreviewFiles()
	if err != nil {
		return 0, "", fmt.Errorf("error during file scanning phase: %w", err)
	}

	outputData := OutputJSON{
		ProjectFiles: []ProcessedFile{},
	}

	if len(filesToProcess) == 0 {
		if progressCallback != nil {
			progressCallback(0, 0)
		}
		jsonBytes, marshalErr := json.MarshalIndent(outputData, "", "  ")
		if marshalErr != nil {
			return 0, "", fmt.Errorf("marshaling empty output JSON: %w", marshalErr)
		}
		if err := os.WriteFile(fc.Config.Output, jsonBytes, 0644); err != nil {
			return 0, "", fmt.Errorf("writing empty output file: %w", err)
		}
		return 0, utils.FormatSize(int64(len(jsonBytes))), nil
	}

	var allProcessedFiles []ProcessedFile
	var wg sync.WaitGroup
	var mu sync.Mutex
	numWorkers := runtime.NumCPU()
	if numWorkers > len(filesToProcess) {
		numWorkers = len(filesToProcess)
	}
	jobs := make(chan FileEntry, len(filesToProcess))
	processedCount := 0

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				processed, err := fc.processFile(entry)
				if err != nil {
					fmt.Printf("Error processing file %s: %v\n", entry.SourcePath, err)
					mu.Lock()
					processedCount++
					if progressCallback != nil {
						progressCallback(processedCount, len(filesToProcess))
					}
					mu.Unlock()
					continue
				}
				if processed != nil {
					mu.Lock()
					allProcessedFiles = append(allProcessedFiles, processed)
					mu.Unlock()
				}
				mu.Lock()
				processedCount++
				if progressCallback != nil {
					progressCallback(processedCount, len(filesToProcess))
				}
				mu.Unlock()
			}
		}()
	}

	for _, fileEntry := range filesToProcess {
		jobs <- fileEntry
	}
	close(jobs)
	wg.Wait()

	if progressCallback != nil {
		progressCallback(len(filesToProcess), len(filesToProcess))
	}

	outputData.ProjectFiles = allProcessedFiles
	jsonBytes, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return 0, "", fmt.Errorf("marshaling JSON output: %w", err)
	}
	err = os.WriteFile(fc.Config.Output, jsonBytes, 0644)
	if err != nil {
		return 0, "", fmt.Errorf("writing output file: %w", err)
	}
	return len(allProcessedFiles), utils.FormatSize(int64(len(jsonBytes))), nil
}

// GetFileContent retrieves the raw content of a file given its original relative path.
func (fc *FileCollector) GetFileContent(originalRelPath string) (string, error) {
	absPath := filepath.Join(fc.Config.Root, originalRelPath)
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", absPath, err)
	}
	return string(contentBytes), nil
}
