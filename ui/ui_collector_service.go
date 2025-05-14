package ui

import (
	"fmt"
	"projectson/collector"
	"projectson/config"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2" // Для типа fyne.App и других
)

// RunStats holds statistics about a collection run.
type RunStats struct {
	FileCount    int
	OutputSize   string
	ProcessTime  time.Duration
	Timestamp    time.Time
	ErrorMessage string
	ConfigUsed   *config.Config
}

// CollectorService manages the collector instance and shared data.
type CollectorService struct {
	config                *config.Config
	fileCollector         *collector.FileCollector
	needsCollectorRebuild bool

	PreviewedFiles []collector.FileEntry
	PreviewError   error

	LastRunStats RunStats
	mu           sync.Mutex

	app          fyne.App // Храним экземпляр приложения
	parentWindow fyne.Window
}

func NewCollectorService(initialConfig *config.Config, appInstance fyne.App, window fyne.Window) *CollectorService {
	if appInstance == nil {
		panic("CollectorService: fyne.App instance cannot be nil for Driver().RunOnMain()")
	}
	cs := &CollectorService{
		config:                initialConfig,
		needsCollectorRebuild: true,
		app:                   appInstance,
		parentWindow:          window,
	}
	return cs
}

// Helper to safely run on main UI thread
func (cs *CollectorService) runTaskOnUITread(fn func()) {
	if cs.app == nil {
		fmt.Println("Error: CollectorService.app is nil, cannot run task on UI thread.")
		return
	}
	driver := cs.app.Driver()
	if driver == nil {
		fmt.Println("Error: CollectorService.app.Driver() is nil, cannot run task on UI thread.")
		return
	}
	driver.DoFromGoroutine(fn, true) // Используем метод драйвера
}

func (cs *CollectorService) SetParentWindow(w fyne.Window) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.parentWindow = w
}

func (cs *CollectorService) ParentWindow() fyne.Window {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.parentWindow == nil {
		fmt.Println("Warning: CollectorService parentWindow not explicitly set!")
		if cs.app != nil {
			if windows := cs.app.Driver().AllWindows(); len(windows) > 0 {
				return windows[0]
			}
		}
	}
	return cs.parentWindow
}

func (cs *CollectorService) rebuildCollectorIfNeeded() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.fileCollector == nil || cs.needsCollectorRebuild {
		if cs.config == nil {
			return fmt.Errorf("cannot build collector: config is nil")
		}
		if err := cs.config.Validate(); err != nil {
			cs.fileCollector = nil
			cs.needsCollectorRebuild = true
			return fmt.Errorf("configuration validation failed: %w", err)
		}

		newCollector, err := collector.NewFileCollector(cs.config)
		if err != nil {
			cs.fileCollector = nil
			cs.needsCollectorRebuild = true
			return fmt.Errorf("failed to create file collector: %w. Check exclude patterns for valid regex.", err)
		}
		cs.fileCollector = newCollector
		cs.needsCollectorRebuild = false
		fmt.Println("FileCollector instance (re)built.")
	}
	return nil
}

func (cs *CollectorService) UpdateConfig(newConfig *config.Config) {
	cs.mu.Lock()
	cs.config = newConfig
	cs.needsCollectorRebuild = true
	cs.mu.Unlock()

	cs.ClearPreview()
	cs.ClearStats()
}

func (cs *CollectorService) GetConfig() *config.Config {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.config == nil {
		return config.NewDefaultConfig()
	}
	cfgCopy := *cs.config
	return &cfgCopy
}

func (cs *CollectorService) GetCurrentFileCollector() (*collector.FileCollector, error) {
	if err := cs.rebuildCollectorIfNeeded(); err != nil {
		return nil, err
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.fileCollector == nil {
		return nil, fmt.Errorf("collector is not initialized; rebuild failed or config is missing")
	}
	return cs.fileCollector, nil
}

func (cs *CollectorService) PerformPreview(onComplete func(files []collector.FileEntry, err error)) {
	go func() {
		currentCollector, err := cs.GetCurrentFileCollector()
		if err != nil {
			cs.runTaskOnUITread(func() { onComplete(nil, err) })
			return
		}

		files, previewErr := currentCollector.PreviewFiles()
		cs.mu.Lock()
		if previewErr != nil {
			cs.PreviewError = previewErr
			cs.PreviewedFiles = nil
		} else {
			cs.PreviewError = nil
			cs.PreviewedFiles = files
		}
		cs.mu.Unlock()
		cs.runTaskOnUITread(func() { onComplete(files, previewErr) })
	}()
}

func (cs *CollectorService) GetPreviewedFiles() ([]collector.FileEntry, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.PreviewedFiles, cs.PreviewError
}

func (cs *CollectorService) ClearPreview() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.PreviewedFiles = nil
	cs.PreviewError = nil
}

func (cs *CollectorService) PerformRun(
	progressCallback func(current, total int),
	onComplete func(stats RunStats),
) {
	go func() {
		currentCollector, err := cs.GetCurrentFileCollector()
		runConfig := cs.GetConfig()

		if err != nil {
			runStats := RunStats{
				ErrorMessage: fmt.Sprintf("Failed to get collector for run: %v", err),
				Timestamp:    time.Now(),
				ConfigUsed:   runConfig,
			}
			cs.mu.Lock()
			cs.LastRunStats = runStats
			cs.mu.Unlock()
			cs.runTaskOnUITread(func() {
				if progressCallback != nil {
					progressCallback(0, 0)
				}
				onComplete(runStats)
			})
			return
		}

		startTime := time.Now()
		uiProgressCallbackWrapper := func(current, total int) {
			cs.runTaskOnUITread(func() {
				if progressCallback != nil {
					progressCallback(current, total)
				}
			})
		}

		count, size, runErr := currentCollector.Run(uiProgressCallbackWrapper)
		endTime := time.Now()

		var runStats RunStats
		if runErr != nil {
			runStats = RunStats{
				ErrorMessage: fmt.Sprintf("Run failed: %v", runErr),
				Timestamp:    endTime,
				ConfigUsed:   runConfig,
			}
		} else {
			runStats = RunStats{
				FileCount:   count,
				OutputSize:  size,
				ProcessTime: endTime.Sub(startTime),
				Timestamp:   endTime,
				ConfigUsed:  runConfig,
			}
		}
		cs.mu.Lock()
		cs.LastRunStats = runStats
		cs.mu.Unlock()

		cs.runTaskOnUITread(func() { onComplete(runStats) })
	}()
}

func (cs *CollectorService) GetLastRunStats() RunStats {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.LastRunStats
}

func (cs *CollectorService) ClearStats() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.LastRunStats = RunStats{}
}

func (cs *CollectorService) GetFileContentWithExclusions(originalRelPath string, fileExt string) (original, modified string, err error) {
	currentCollector, collectorErr := cs.GetCurrentFileCollector()
	if collectorErr != nil {
		return "", "", collectorErr
	}

	original, err = currentCollector.GetFileContent(originalRelPath)
	if err != nil {
		return "", "", err
	}
	modified, err = currentCollector.ApplyContentExclusions(original, fileExt)
	if err != nil {
		return original, "", fmt.Errorf("error applying exclusions: %w", err)
	}
	return original, modified, nil
}

func CleanSplit(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	if len(cleaned) == 0 {
		return []string{}
	}
	return cleaned
}
