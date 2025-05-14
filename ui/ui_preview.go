package ui

import (
	"fmt"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"projectson/collector" // For FileEntry
	"projectson/utils"     // For FormatSize
)

// MakePreviewPage creates the UI for previewing files.
func MakePreviewPage(collectorService *CollectorService, window fyne.Window) fyne.CanvasObject {
	var currentFiles []collector.FileEntry
	statusLabel := widget.NewLabel("Press 'Refresh File List' to scan for files.")

	fileList := widget.NewList(
		func() int {
			return len(currentFiles)
		},
		func() fyne.CanvasObject {
			pathLabel := widget.NewLabel("path/to/very/very/long/file/example/that/should/wrap.txt")
			pathLabel.Wrapping = fyne.TextWrapWord

			formatLabel := widget.NewLabel("fmt")
			modeLabel := widget.NewLabel("mode")
			sizeLabel := widget.NewLabel("size")

			detailsLine := container.NewHBox(
				widget.NewLabel("Format:"), formatLabel,
				layout.NewSpacer(),
				widget.NewLabel("Mode:"), modeLabel,
				layout.NewSpacer(),
				widget.NewLabel("Size:"), sizeLabel,
			)
			return container.NewVBox(pathLabel, detailsLine)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < 0 || id >= len(currentFiles) {
				return
			}
			entry := currentFiles[id]
			vBox := item.(*fyne.Container)

			pathLabel := vBox.Objects[0].(*widget.Label)
			pathLabel.SetText(entry.Path)

			detailsLine := vBox.Objects[1].(*fyne.Container)
			detailsHBox := detailsLine
			detailsHBox.Objects[1].(*widget.Label).SetText(entry.Format)
			detailsHBox.Objects[4].(*widget.Label).SetText(entry.Mode)
			detailsHBox.Objects[7].(*widget.Label).SetText(utils.FormatSize(entry.Size))
		},
	)

	originalContent := widget.NewMultiLineEntry()
	originalContent.Wrapping = fyne.TextWrapOff
	originalContent.Disable()                                   // Corrected: Use SetReadOnly method
	originalContent.TextStyle = fyne.TextStyle{Monospace: true} // For code-like appearance

	modifiedContent := widget.NewMultiLineEntry()
	modifiedContent.Wrapping = fyne.TextWrapOff
	modifiedContent.Disable()                                   // Corrected: Use SetReadOnly method
	modifiedContent.TextStyle = fyne.TextStyle{Monospace: true} // For code-like appearance

	originalCard := widget.NewCard("Original Content", "", container.NewScroll(originalContent))
	modifiedCard := widget.NewCard("Content After Exclusions", "", container.NewScroll(modifiedContent))

	fileList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(currentFiles) {
			originalContent.SetText("")
			modifiedContent.SetText("")
			originalCard.SetTitle("Original Content")
			modifiedCard.SetTitle("Content After Exclusions")
			return
		}
		selected := currentFiles[id]
		baseName := filepath.Base(selected.Path)
		originalCard.SetTitle(fmt.Sprintf("Original: %s", baseName))
		modifiedCard.SetTitle(fmt.Sprintf("Modified: %s", baseName))

		origText, modText, err := collectorService.GetFileContentWithExclusions(selected.OriginalPath, selected.Format)
		if err != nil {
			dialog.ShowError(fmt.Errorf("error getting content for %s: %w", selected.Path, err), window)
			originalContent.SetText("Error loading content.")
			modifiedContent.SetText("Error loading content.")
			return
		}
		originalContent.SetText(origText)
		modifiedContent.SetText(modText)
	}

	formatFilter := widget.NewSelect([]string{"All"}, nil)
	modeFilter := widget.NewSelect([]string{"All"}, nil)

	applyFilters := func() {
		allPreviewedFiles, _ := collectorService.GetPreviewedFiles()
		if allPreviewedFiles == nil {
			currentFiles = []collector.FileEntry{}
			fileList.Refresh()
			statusLabel.SetText(fmt.Sprintf("Files: 0"))
			return
		}

		var filtered []collector.FileEntry
		selFormat := formatFilter.Selected
		selMode := modeFilter.Selected

		for _, f := range allPreviewedFiles {
			formatMatch := (selFormat == "All" || f.Format == selFormat)
			modeMatch := (selMode == "All" || f.Mode == selMode)
			if formatMatch && modeMatch {
				filtered = append(filtered, f)
			}
		}
		currentFiles = filtered
		sort.Slice(currentFiles, func(i, j int) bool {
			return currentFiles[i].Path < currentFiles[j].Path
		})
		fileList.Refresh()
		fileList.UnselectAll()
		originalContent.SetText("")
		modifiedContent.SetText("")
		originalCard.SetTitle("Original Content")
		modifiedCard.SetTitle("Content After Exclusions")
		statusLabel.SetText(fmt.Sprintf("Files: %d (filtered from %d)", len(currentFiles), len(allPreviewedFiles)))
	}

	formatFilter.OnChanged = func(s string) { applyFilters() }
	modeFilter.OnChanged = func(s string) { applyFilters() }

	populateFilters := func(files []collector.FileEntry) {
		if files == nil {
			formatFilter.Options = []string{"All"}
			modeFilter.Options = []string{"All"}
			formatFilter.SetSelected("All")
			modeFilter.SetSelected("All")
			return
		}
		formats := map[string]bool{"All": true}
		modes := map[string]bool{"All": true}
		for _, f := range files {
			formats[f.Format] = true
			modes[f.Mode] = true
		}

		var formatOpts, modeOpts []string
		for f := range formats {
			formatOpts = append(formatOpts, f)
		}
		for m := range modes {
			modeOpts = append(modeOpts, m)
		}
		sort.Strings(formatOpts)
		sort.Strings(modeOpts)

		formatFilter.Options = formatOpts
		modeFilter.Options = modeOpts
		if !contains(formatOpts, formatFilter.Selected) {
			formatFilter.SetSelected("All")
		}
		if !contains(modeOpts, modeFilter.Selected) {
			modeFilter.SetSelected("All")
		}
		formatFilter.Refresh()
		modeFilter.Refresh()
	}

	loading := widget.NewProgressBarInfinite()
	loading.Hide()
	var refreshButton *widget.Button

	refreshButton = widget.NewButtonWithIcon("Refresh File List", theme.ViewRefreshIcon(), func() {
		statusLabel.SetText("Scanning files...")
		loading.Show()
		refreshButton.Disable()

		collectorService.PerformPreview(func(files []collector.FileEntry, err error) {
			loading.Hide()
			refreshButton.Enable()
			if err != nil {
				dialog.ShowError(fmt.Errorf("error previewing files: %w", err), window)
				statusLabel.SetText("Error scanning files.")
				currentFiles = []collector.FileEntry{}
				populateFilters(nil)
			} else {
				statusLabel.SetText(fmt.Sprintf("Found %d files.", len(files)))
				populateFilters(files)
			}
			applyFilters()
		})
	})

	contentSplit := container.NewHSplit(
		container.NewMax(originalCard),
		container.NewMax(modifiedCard),
	)
	contentSplit.Offset = 0.5

	filterToolbar := container.NewHBox(
		widget.NewLabel("Filter by Format:"), formatFilter,
		widget.NewLabel("Mode:"), modeFilter,
	)

	return container.NewBorder(
		container.NewVBox(
			container.NewHBox(refreshButton, loading),
			statusLabel,
			filterToolbar,
		),
		nil,
		nil,
		nil,
		container.NewVSplit(
			fileList,
			contentSplit,
		),
	)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
