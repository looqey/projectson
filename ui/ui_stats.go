package ui

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// MakeStatsPage creates the UI for displaying statistics.
func MakeStatsPage(collectorService *CollectorService) fyne.CanvasObject {
	stats := collectorService.GetLastRunStats()
	previewFiles, _ := collectorService.GetPreviewedFiles() // Get files from last preview for distribution charts

	if stats.Timestamp.IsZero() && (previewFiles == nil || len(previewFiles) == 0) {
		return container.NewCenter(widget.NewLabel("No statistics available. Run the collection process or Preview files first."))
	}

	mainVBox := container.NewVBox()

	// Last Run Statistics
	if !stats.Timestamp.IsZero() {
		mainVBox.Add(widget.NewLabelWithStyle("Last Collection Run Statistics:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

		runDetails := container.New(layout.NewFormLayout(),
			widget.NewLabel("Files Processed:"), widget.NewLabel(fmt.Sprintf("%d", stats.FileCount)),
			widget.NewLabel("Output Size:"), widget.NewLabel(stats.OutputSize),
			widget.NewLabel("Processing Time:"), widget.NewLabel(fmt.Sprintf("%.2f seconds", stats.ProcessTime.Seconds())),
			widget.NewLabel("Timestamp:"), widget.NewLabel(stats.Timestamp.Format("2006-01-02 15:04:05")),
		)
		if stats.ErrorMessage != "" {
			runDetails.Add(widget.NewLabel("Status:"))
			runDetails.Add(widget.NewLabel("Failed: " + stats.ErrorMessage))
		}
		mainVBox.Add(runDetails)
		mainVBox.Add(widget.NewSeparator())
	}

	// File Distribution (from previewed files)
	if previewFiles != nil && len(previewFiles) > 0 {
		mainVBox.Add(widget.NewLabelWithStyle("File Distribution (from last Preview):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

		// Format Distribution
		formatCounts := make(map[string]int)
		for _, f := range previewFiles {
			formatCounts[f.Format]++
		}
		var formatStrings []string
		for format, count := range formatCounts {
			formatStrings = append(formatStrings, fmt.Sprintf("%s: %d", format, count))
		}
		sort.Strings(formatStrings) // Sort for consistent display
		formatCard := widget.NewCard("Files by Format", "", widget.NewLabel(strings.Join(formatStrings, "\n")))

		// Mode Distribution
		modeCounts := make(map[string]int)
		for _, f := range previewFiles {
			modeCounts[f.Mode]++
		}
		var modeStrings []string
		for mode, count := range modeCounts {
			modeStrings = append(modeStrings, fmt.Sprintf("%s: %d", mode, count))
		}
		sort.Strings(modeStrings)
		modeCard := widget.NewCard("Files by Collection Mode", "", widget.NewLabel(strings.Join(modeStrings, "\n")))

		// Size Distribution
		sizeCategories := map[string]int{
			"< 1KB":     0,
			"1-10KB":    0,
			"10-100KB":  0,
			"100KB-1MB": 0,
			"> 1MB":     0,
		}
		for _, f := range previewFiles {
			switch {
			case f.Size < 1024:
				sizeCategories["< 1KB"]++
			case f.Size < 10*1024:
				sizeCategories["1-10KB"]++
			case f.Size < 100*1024:
				sizeCategories["10-100KB"]++
			case f.Size < 1024*1024:
				sizeCategories["100KB-1MB"]++
			default:
				sizeCategories["> 1MB"]++
			}
		}
		var sizeStrings []string
		// Ordered display for size categories
		orderedSizeKeys := []string{"< 1KB", "1-10KB", "10-100KB", "100KB-1MB", "> 1MB"}
		for _, key := range orderedSizeKeys {
			sizeStrings = append(sizeStrings, fmt.Sprintf("%s: %d", key, sizeCategories[key]))
		}
		sizeCard := widget.NewCard("Files by Size Category", "", widget.NewLabel(strings.Join(sizeStrings, "\n")))

		distGrid := container.NewGridWithColumns(3, formatCard, modeCard, sizeCard)
		mainVBox.Add(distGrid)
		mainVBox.Add(widget.NewSeparator())
	}

	// Content Exclusion Rules (from current config)
	// Or use stats.ConfigUsed if you want stats tied to config used for THAT specific run.
	// For now, show currently active config's exclusion rules.
	currentCfg := collectorService.GetConfig()
	if currentCfg.ContentExclusions != nil && len(currentCfg.ContentExclusions) > 0 {
		mainVBox.Add(widget.NewLabelWithStyle("Active Content Exclusion Rules:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))

		exclusionTable := container.NewVBox() // Using VBox to list rules
		header := container.NewGridWithColumns(3,
			widget.NewLabelWithStyle("Type", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("File Pattern", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		)
		exclusionTable.Add(header)

		for _, rule := range currentCfg.ContentExclusions {
			details := ""
			if rule.Type == "delimiters" {
				details = fmt.Sprintf("Start: '%s', End: '%s'", rule.Start, rule.End)
			} else {
				details = fmt.Sprintf("Regex: '%s'", rule.Pattern)
			}
			row := container.NewGridWithColumns(3,
				widget.NewLabel(strings.Title(rule.Type)), // Title case for "Delimiters", "Regexp"
				widget.NewLabel(rule.FilePattern),
				widget.NewLabel(details),
			)
			exclusionTable.Add(row)
		}
		mainVBox.Add(widget.NewCard("", "", exclusionTable)) // Wrap in card for bordering/padding
	}

	return container.NewVScroll(mainVBox)
}
