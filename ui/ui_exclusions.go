package ui

import (
	"fmt"
	"strings"

	"projectson/config"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MakeExclusionsPage creates the UI for content exclusion rules.
func MakeExclusionsPage(collectorService *CollectorService, onConfigModified func()) fyne.CanvasObject {
	cfg := collectorService.GetConfig()
	rulesContainer := container.NewVBox()

	applyChangesAndNotify := func() {
		collectorService.UpdateConfig(cfg)
		onConfigModified()
	}

	var rebuildRulesUI func()
	rebuildRulesUI = func() {
		rulesContainer.Objects = nil
		if cfg.ContentExclusions == nil {
			cfg.ContentExclusions = []config.ContentExclusionRule{}
		}
		for i, rule := range cfg.ContentExclusions {
			localIdx := i
			currentRule := rule

			ruleTypeSelect := widget.NewSelect([]string{"delimiters", "regexp"}, nil)
			ruleTypeSelect.SetSelected(currentRule.Type)
			if currentRule.Type == "" {
				ruleTypeSelect.SetSelected("delimiters")
			}

			filePatternEntry := widget.NewEntry()
			filePatternEntry.SetPlaceHolder("e.g., *.vue, * (all files)")
			filePatternEntry.SetText(currentRule.FilePattern)
			if currentRule.FilePattern == "" {
				filePatternEntry.SetText("*")
			}
			filePatternEntry.OnChanged = func(s string) {
				if localIdx < len(cfg.ContentExclusions) {
					cfg.ContentExclusions[localIdx].FilePattern = strings.TrimSpace(s)
					applyChangesAndNotify()
				}
			}

			startDelimiterEntry := widget.NewEntry()
			startDelimiterEntry.SetPlaceHolder("<style>")
			startDelimiterEntry.SetText(currentRule.Start)
			startDelimiterEntry.OnChanged = func(s string) {
				if localIdx < len(cfg.ContentExclusions) {
					cfg.ContentExclusions[localIdx].Start = s
					applyChangesAndNotify()
				}
			}

			endDelimiterEntry := widget.NewEntry()
			endDelimiterEntry.SetPlaceHolder("</style>")
			endDelimiterEntry.SetText(currentRule.End)
			endDelimiterEntry.OnChanged = func(s string) {
				if localIdx < len(cfg.ContentExclusions) {
					cfg.ContentExclusions[localIdx].End = s
					applyChangesAndNotify()
				}
			}

			regexEntry := widget.NewEntry()
			regexEntry.SetPlaceHolder("<style>[\\s\\S]*?</style>")
			regexEntry.SetText(currentRule.Pattern)
			regexEntry.OnChanged = func(s string) {
				if localIdx < len(cfg.ContentExclusions) {
					cfg.ContentExclusions[localIdx].Pattern = s
					applyChangesAndNotify()
				}
			}

			delimiterFields := widget.NewForm(
				widget.NewFormItem("Start Delimiter", startDelimiterEntry),
				widget.NewFormItem("End Delimiter", endDelimiterEntry),
			)
			regexpFormItemContainer := container.NewVBox(widget.NewForm(widget.NewFormItem("Regular Expression", regexEntry)))
			specificsContainer := container.NewStack()

			updateSpecificsVisibility := func(ruleType string) {
				if ruleType == "delimiters" {
					specificsContainer.Objects = []fyne.CanvasObject{delimiterFields}
				} else { // regexp
					specificsContainer.Objects = []fyne.CanvasObject{regexpFormItemContainer}
				}
				specificsContainer.Refresh()
			}

			ruleTypeSelect.OnChanged = func(selectedType string) {
				if localIdx < len(cfg.ContentExclusions) {
					cfg.ContentExclusions[localIdx].Type = selectedType
					updateSpecificsVisibility(selectedType) // Update UI for fields
					applyChangesAndNotify()                 // Apply the type change
				}
			}
			updateSpecificsVisibility(ruleTypeSelect.Selected) // Initial setup

			deleteRuleButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				if localIdx < len(cfg.ContentExclusions) {
					cfg.ContentExclusions = append(cfg.ContentExclusions[:localIdx], cfg.ContentExclusions[localIdx+1:]...)
					rebuildRulesUI()        // Rebuild the list
					applyChangesAndNotify() // Apply deletion
				}
			})

			ruleCardContent := container.NewVBox(
				widget.NewForm(
					widget.NewFormItem("Type", ruleTypeSelect),
					widget.NewFormItem("File Pattern", filePatternEntry),
				),
				specificsContainer,
			)
			rulePresentation := container.NewBorder(
				nil, nil, nil, deleteRuleButton,
				widget.NewCard(fmt.Sprintf("Rule #%d", localIdx+1), "", ruleCardContent),
			)
			rulesContainer.Add(rulePresentation)
		}
		rulesContainer.Refresh()
	}
	rebuildRulesUI()

	addNewRuleButton := widget.NewButtonWithIcon("Add New Exclusion Rule", theme.ContentAddIcon(), func() {
		newRule := config.ContentExclusionRule{Type: "delimiters", FilePattern: "*"}
		cfg.ContentExclusions = append(cfg.ContentExclusions, newRule)
		rebuildRulesUI()
		applyChangesAndNotify()
	})

	helpText := `Content Exclusions help remove unwanted sections from files before processing.
Types:
- Delimiters: Define start and end text tags (e.g., <script> and </script>).
- Regexp: Use regular expressions for complex patterns.
File Pattern:
- Glob pattern to match files (e.g., "*.vue", "main.js", "*" for all).
- Matches against file extension (e.g., "vue") or full filename (e.g. ".env.example").
Examples:
1. Exclude Vue <style> blocks:
   Type: delimiters, File Pattern: *.vue
   Start: <style>, End: </style>
   (Or use regexp with pattern: <style.*?>[\s\S]*?</style.*?> for attributes)
2. Exclude JS comments:
   Type: regexp, File Pattern: *.js
   Pattern: (//.*)|(/\*[\s\S]*?\*/)`
	helpCard := widget.NewCard("Help & Examples", "", widget.NewLabel(helpText))

	return container.NewVScroll(container.NewVBox(
		rulesContainer,
		addNewRuleButton,
		widget.NewSeparator(),
		helpCard,
	))
}
