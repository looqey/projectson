package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage" // For Download button
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"os"
	"path/filepath"
	"strings"
)

// MakeRunPage creates the UI for running the collection process.
// onRunComplete is a callback to signal main app to e.g. refresh stats tab
func MakeRunPage(collectorService *CollectorService, window fyne.Window, statusBar *widget.Label, onRunComplete func()) fyne.CanvasObject {
	cfg := collectorService.GetConfig() // Get current config for display

	// Configuration Summary (display-only)
	// These labels will need to be updated if config changes and page is rebuilt.
	rootLabel := widget.NewLabel("Root: " + cfg.Root)
	formatsLabel := widget.NewLabel("Formats: " + strings.Join(cfg.Formats, ", "))
	outputLabel := widget.NewLabel("Output: " + cfg.Output)
	includesLabel := widget.NewLabel(fmt.Sprintf("Includes: %d rules", len(cfg.Include)))
	excludesLabel := widget.NewLabel(fmt.Sprintf("Excludes: %d patterns", len(cfg.ExcludePatterns)))
	contentExclLabel := widget.NewLabel(fmt.Sprintf("Content Exclusions: %d rules", len(cfg.ContentExclusions)))

	configSummaryBox := container.NewVBox(
		widget.NewLabelWithStyle("Current Configuration Summary:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		rootLabel, formatsLabel, outputLabel, includesLabel, excludesLabel, contentExclLabel,
	)

	filesToProcessLabel := widget.NewLabel("Files to process: (Run Preview first or it will scan on Run)")
	// Update filesToProcessLabel if preview data is available
	previewFiles, _ := collectorService.GetPreviewedFiles()
	if previewFiles != nil {
		filesToProcessLabel.SetText(fmt.Sprintf("Files to process (from last preview): %d", len(previewFiles)))
	}

	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	progressStatus := widget.NewLabel("")
	progressStatus.Hide()

	runButton := widget.NewButtonWithIcon("Run Collection Process", theme.MediaPlayIcon(), nil) // Action set later
	downloadButton := widget.NewButtonWithIcon("Save Output JSON", theme.DownloadIcon(), nil)
	downloadButton.Disable() // Enabled if output file exists

	checkOutputFile := func() {
		currentConfig := collectorService.GetConfig() // Get fresh config for output path
		if _, err := os.Stat(currentConfig.Output); err == nil {
			downloadButton.Enable()
		} else {
			downloadButton.Disable()
		}
	}
	checkOutputFile() // Initial check

	downloadButton.OnTapped = func() {
		currentConfig := collectorService.GetConfig() // Get fresh config for output path
		storage.NewFileURI(currentConfig.Output)

		// Create a file save dialog to allow user to choose where to "download" (save a copy)
		saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if writer == nil { // Cancelled
				return
			}
			defer writer.Close()

			sourceData, readErr := os.ReadFile(currentConfig.Output)
			if readErr != nil {
				dialog.ShowError(fmt.Errorf("failed to read original output file: %w", readErr), window)
				return
			}
			_, writeErr := writer.Write(sourceData)
			if writeErr != nil {
				dialog.ShowError(fmt.Errorf("failed to save copy of output file: %w", writeErr), window)
				return
			}
			dialog.ShowInformation("Download Complete", "Output JSON saved to: "+writer.URI().Path(), window)
		}, window)
		saveDialog.SetFileName(filepath.Base(currentConfig.Output)) // Suggest original filename
		saveDialog.Show()
	}

	runButton.OnTapped = func() {
		runButton.Disable()
		progressBar.SetValue(0)
		progressBar.Show()
		progressStatus.SetText("Preparing to run...")
		progressStatus.Show()
		statusBar.SetText("Processing files...") // Update main status bar too

		// Ensure collector is up-to-date before run (important if user didn't click "Validate" on config page)
		if err := collectorService.rebuildCollectorIfNeeded(); err != nil {
			dialog.ShowError(fmt.Errorf("failed to initialize collector before run: %w. Check config and validate.", err), window)
			runButton.Enable()
			progressBar.Hide()
			progressStatus.Hide()
			statusBar.SetText("Run failed: Collector initialization error.")
			return
		}

		collectorService.PerformRun(
			func(current, total int) { // Progress callback
				if total > 0 {
					progressBar.Max = float64(total)
					progressBar.SetValue(float64(current))
					progressStatus.SetText(fmt.Sprintf("Processing: %d / %d", current, total))
				} else {
					progressStatus.SetText(fmt.Sprintf("Processing: %d (total unknown yet)", current))
				}
			},
			func(stats RunStats) { // OnComplete callback
				runButton.Enable()
				progressBar.Hide() // Or set to 100% then hide
				progressStatus.Hide()

				if stats.ErrorMessage != "" {
					dialog.ShowError(fmt.Errorf(stats.ErrorMessage), window)
					statusBar.SetText("Run failed: " + stats.ErrorMessage)
				} else {
					successMsg := fmt.Sprintf("✅ Successfully processed %d files in %.2f seconds. Output size: %s",
						stats.FileCount, stats.ProcessTime.Seconds(), stats.OutputSize)
					dialog.ShowInformation("Run Complete", successMsg, window)
					statusBar.SetText(successMsg)
					filesToProcessLabel.SetText(fmt.Sprintf("Files processed in last run: %d", stats.FileCount))
				}
				checkOutputFile() // Check if output file exists to enable download button
				if onRunComplete != nil {
					onRunComplete() // Notify main app, e.g., to refresh stats tab
				}
			},
		)
	}

	// The UI should be rebuilt if config changes to update the summary.
	// This can be done by the main app's `fullUIUpdateOnConfigChange`.
	// So, this MakeRunPage itself doesn't need an `onConfigModified` callback.

	return container.NewVScroll(container.NewVBox(
		configSummaryBox,
		widget.NewSeparator(),
		filesToProcessLabel,
		widget.NewSeparator(),
		runButton,
		progressBar,
		progressStatus,
		widget.NewSeparator(),
		downloadButton,
	))
}
