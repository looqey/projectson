package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MakeConfigPage creates the UI for the configuration settings.
func MakeConfigPage(collectorService *CollectorService, onConfigModified func()) fyne.CanvasObject {
	cfg := collectorService.GetConfig()
	parentWin := collectorService.ParentWindow()

	rootHelp := "Specify the root directory of your project."
	formatsHelp := "List file extensions to include (one per line, e.g., 'vue', 'ts')."
	outputHelp := "Specify the full path for the output JSON file."
	includesHelp := "Paths to include, relative to Project Root. Syntax: path[:mode] or path/*[:mode]. Modes: path, content, both (default). '/*' means non-recursive."
	excludesHelp := "Patterns to exclude files/directories (one per line). Glob (e.g., node_modules, *.log) or /regex/."

	applyChangesAndNotify := func() {
		// Basic validation before updating service
		// More comprehensive validation (like regex compile check) happens when collector is built
		if err := cfg.Validate(); err != nil {
			dialog.ShowError(fmt.Errorf("Configuration validation failed: %w\nApplying anyway, but collector may fail.", err), parentWin)
			// Proceed to update, errors will be caught by collector use
		}
		// Test collector creation can be slow for every keystroke.
		// It will be validated when collector is actually needed (Preview/Run).
		// _, testCollectorErr := collector.NewFileCollector(cfg)
		// if testCollectorErr != nil {
		//    dialog.ShowError(fmt.Errorf("Configuration error (e.g., invalid regex): %w", testCollectorErr), parentWin)
		//    // return // Optionally prevent apply if there's an immediate structural error
		// }

		collectorService.UpdateConfig(cfg)
		onConfigModified()
	}

	rootEntry := widget.NewEntry()
	rootEntry.SetText(cfg.Root)
	rootEntry.OnChanged = func(s string) {
		cfg.Root = strings.TrimSpace(s)
		applyChangesAndNotify()
	}
	browseRootButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, parentWin)
				return
			}
			if uri != nil {
				rootEntry.SetText(uri.Path()) // Triggers OnChanged
			}
		}, parentWin)
	})
	rootContainer := container.NewBorder(nil, nil, nil, browseRootButton, rootEntry)

	formatsEntry := widget.NewMultiLineEntry()
	formatsEntry.SetPlaceHolder("e.g.\nvue\nts\njs")
	formatsEntry.SetText(strings.Join(cfg.Formats, "\n"))
	formatsEntry.OnChanged = func(s string) {
		cfg.Formats = CleanSplit(s)
		applyChangesAndNotify()
	}
	formatsEntry.Wrapping = fyne.TextWrapOff
	formatsEntry.SetMinRowsVisible(3)

	outputEntry := widget.NewEntry()
	outputEntry.SetText(cfg.Output)
	outputEntry.OnChanged = func(s string) {
		cfg.Output = strings.TrimSpace(s)
		applyChangesAndNotify()
	}
	browseOutputButton := widget.NewButtonWithIcon("Browse", theme.FileIcon(), func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, parentWin)
				return
			}
			if writer != nil {
				outputEntry.SetText(writer.URI().Path()) // Triggers OnChanged
				_ = writer.Close()
			}
		}, parentWin)
	})
	outputContainer := container.NewBorder(nil, nil, nil, browseOutputButton, outputEntry)

	includesListContainer := container.NewVBox()
	var rebuildIncludesUI func()
	rebuildIncludesUI = func() {
		includesListContainer.Objects = nil
		if cfg.Include == nil {
			cfg.Include = []string{}
		}
		for i, incFullString := range cfg.Include {
			localIdx := i
			pathForEntry := incFullString
			modeForSelect := "both"
			colonIndex := strings.LastIndex(incFullString, ":")
			if colonIndex != -1 {
				pathSegment := strings.TrimSpace(incFullString[:colonIndex])
				modeSegment := strings.ToLower(strings.TrimSpace(incFullString[colonIndex+1:]))
				if modeSegment == "path" || modeSegment == "content" {
					pathForEntry = pathSegment
					modeForSelect = modeSegment
				} else if modeSegment == "both" {
					pathForEntry = pathSegment
				}
			}

			pathEntryItem := widget.NewEntry()
			pathEntryItem.SetText(pathForEntry)
			pathEntryItem.SetPlaceHolder("path/to/include/*:mode")

			modeSelectItem := widget.NewSelect([]string{"both", "path", "content"}, nil)
			modeSelectItem.SetSelected(modeForSelect)

			updateIncludeEntry := func() {
				currentPathInput := strings.TrimSpace(pathEntryItem.Text)
				currentModeSelected := modeSelectItem.Selected
				finalEntryString := ""
				if currentPathInput == "" {
					finalEntryString = ""
				} else if currentModeSelected != "both" {
					finalEntryString = currentPathInput + ":" + currentModeSelected
				} else {
					finalEntryString = currentPathInput
				}
				if localIdx < len(cfg.Include) {
					cfg.Include[localIdx] = finalEntryString
					applyChangesAndNotify()
				}
			}
			pathEntryItem.OnChanged = func(s string) { updateIncludeEntry() }
			modeSelectItem.OnChanged = func(s string) { updateIncludeEntry() }

			removeButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				if localIdx < len(cfg.Include) {
					cfg.Include = append(cfg.Include[:localIdx], cfg.Include[localIdx+1:]...)
					rebuildIncludesUI()
					applyChangesAndNotify()
				}
			})
			entryRow := container.NewBorder(nil, nil, removeButton, nil, container.NewGridWithColumns(2, pathEntryItem, modeSelectItem))
			includesListContainer.Add(entryRow)
		}
		includesListContainer.Refresh()
	}
	rebuildIncludesUI()

	addIncludeButton := widget.NewButtonWithIcon("Add Include Path", theme.ContentAddIcon(), func() {
		cfg.Include = append(cfg.Include, "new_path:both") // Default new entry
		rebuildIncludesUI()
		applyChangesAndNotify()
	})

	excludePatternsEntry := widget.NewMultiLineEntry()
	excludePatternsEntry.SetPlaceHolder("e.g.\nnode_modules\n*.min.js\n/test_.+/ (regex)")
	excludePatternsEntry.SetText(strings.Join(cfg.ExcludePatterns, "\n"))
	excludePatternsEntry.OnChanged = func(s string) {
		cfg.ExcludePatterns = CleanSplit(s)
		applyChangesAndNotify()
	}
	excludePatternsEntry.Wrapping = fyne.TextWrapOff
	excludePatternsEntry.SetMinRowsVisible(3)

	formItems := []*widget.FormItem{
		newFormFieldWithHelp("Project Root Path", rootContainer, rootHelp, parentWin),
		newFormFieldWithHelp("File Formats (one per line)", formatsEntry, formatsHelp, parentWin),
		newFormFieldWithHelp("Output JSON Path", outputContainer, outputHelp, parentWin),
	}
	baseForm := widget.NewForm(formItems...)

	includesSectionTitle := newLabelWithHelp("Include Paths & Modes", fyne.TextStyle{Bold: true}, includesHelp, parentWin)
	includesSection := container.NewVBox(includesSectionTitle, includesListContainer, addIncludeButton)

	excludesSectionTitle := newLabelWithHelp("Exclude Patterns (one per line, /regex/ or glob)", fyne.TextStyle{Bold: true}, excludesHelp, parentWin)
	excludesSection := container.NewVBox(excludesSectionTitle, excludePatternsEntry)

	return container.NewVScroll(container.NewVBox(
		baseForm,
		widget.NewSeparator(),
		includesSection,
		widget.NewSeparator(),
		excludesSection,
	))
}
