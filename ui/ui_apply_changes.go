package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"projectson/collector" // For AIResponse and AIFileModification
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func MakeApplyChangesPage(collectorService *CollectorService, window fyne.Window, statusBar *widget.Label) fyne.CanvasObject {
	aiResponseEntry := widget.NewMultiLineEntry()
	aiResponseEntry.SetPlaceHolder("Paste AI-generated JSON response here...")
	aiResponseEntry.Wrapping = fyne.TextWrapOff // JSON is often better without wrapping
	aiResponseEntry.SetMinRowsVisible(15)

	// The AI response 'path' field MUST match `FileEntry.Path` format,
	// which is `basename(Root) + os.PathSeparator + relative_path_from_root`.
	// The `applyFileModification` function will need to strip `basename(Root) + os.PathSeparator`
	// to get the `OriginalPath` that can be joined with `Config.Root` to get the absolute file path.

	applyButton := widget.NewButtonWithIcon("Apply Changes to Files", theme.ConfirmIcon(), func() {
		statusBar.SetText("Processing AI response...")
		aiJsonInput := aiResponseEntry.Text
		if strings.TrimSpace(aiJsonInput) == "" {
			dialog.ShowError(fmt.Errorf("AI response input is empty"), window)
			statusBar.SetText("AI response empty.")
			return
		}

		var aiResp collector.AIResponse
		err := json.Unmarshal([]byte(aiJsonInput), &aiResp)
		if err != nil {
			dialog.ShowError(fmt.Errorf("invalid AI JSON response: %w\n\nCheck for syntax errors, unescaped newlines, or incorrect structure. The expected root is {\"modified_files\": [...]}.", err), window)
			statusBar.SetText("Invalid AI JSON.")
			return
		}

		if len(aiResp.ModifiedFiles) == 0 {
			dialog.ShowInformation("No Changes", "AI response contained no files to modify.", window)
			statusBar.SetText("No changes from AI.")
			return
		}

		// Confirmation dialog
		confirmMessage := fmt.Sprintf("You are about to apply %d file modification(s) to your project based on the AI response. This action can overwrite or delete files.\n\nRoot directory: %s\n\nAre you sure you want to proceed? It is recommended to have a backup or use version control.", len(aiResp.ModifiedFiles), collectorService.GetConfig().Root)
		dialog.ShowConfirm("Confirm File Modifications", confirmMessage, func(confirm bool) {
			if !confirm {
				statusBar.SetText("File modification cancelled.")
				return
			}

			statusBar.SetText(fmt.Sprintf("Applying %d changes...", len(aiResp.ModifiedFiles)))
			appliedCount := 0
			errorCount := 0
			var errorMessages []string

			currentConfig := collectorService.GetConfig()
			projectRoot := currentConfig.Root
			rootBasename := filepath.Base(projectRoot) // e.g., "myproject"

			for _, mod := range aiResp.ModifiedFiles {
				// The 'mod.Path' from AI should be like "myproject/src/file.go"
				// We need to get "src/file.go" (the OriginalPath)
				var originalRelativePath string

				// Normalize AI path to use OS-specific separators for prefix checking
				aiPathForCheck := filepath.FromSlash(mod.Path)
				expectedPrefix := rootBasename + string(os.PathSeparator)

				if strings.HasPrefix(aiPathForCheck, expectedPrefix) {
					originalRelativePath = strings.TrimPrefix(aiPathForCheck, expectedPrefix)
				} else {
					// If AI didn't include rootBasename, or used wrong separator not caught by FromSlash
					// (e.g. AI gave "myproject\\src\\file.go" on Linux, FromSlash won't change it)
					// Fallback: try trimming with a generic separator as well if first check fails
					// This is a bit of a safety net. The prompt strongly encourages rootBasename/path format.
					genericPrefixUnix := rootBasename + "/"
					if strings.HasPrefix(mod.Path, genericPrefixUnix) {
						originalRelativePath = strings.TrimPrefix(mod.Path, genericPrefixUnix)
					} else {
						// If AI didn't include rootBasename, assume mod.Path is already the relative path.
						// This is a fallback and might be risky if paths are ambiguous.
						fmt.Printf("Warning: AI path '%s' does not start with project root basename '%s%c'. Assuming it's a direct relative path.\n", mod.Path, rootBasename, os.PathSeparator)
						originalRelativePath = mod.Path
					}
				}

				// Ensure originalRelativePath uses OS-specific separators for joining with projectRoot
				originalRelativePath = filepath.FromSlash(originalRelativePath)
				absPath := filepath.Join(projectRoot, originalRelativePath)

				switch strings.ToLower(mod.Action) {
				case "update":
					// Ensure directory exists
					dir := filepath.Dir(absPath)
					if err := os.MkdirAll(dir, 0755); err != nil {
						errMsg := fmt.Sprintf("Failed to create directory for %s: %v", mod.Path, err)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
						continue
					}
					err := os.WriteFile(absPath, []byte(mod.Content), 0644)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to update %s: %v", mod.Path, err)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
					} else {
						appliedCount++
					}
				case "create":
					// Ensure directory exists
					dir := filepath.Dir(absPath)
					if err := os.MkdirAll(dir, 0755); err != nil {
						errMsg := fmt.Sprintf("Failed to create directory for %s: %v", mod.Path, err)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
						continue
					}
					// Check if file already exists to prevent accidental overwrite with 'create'
					if _, statErr := os.Stat(absPath); statErr == nil {
						errMsg := fmt.Sprintf("File %s already exists. Use 'update' action to overwrite.", mod.Path)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
						continue
					}
					err := os.WriteFile(absPath, []byte(mod.Content), 0644)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to create %s: %v", mod.Path, err)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
					} else {
						appliedCount++
					}
				case "delete":
					err := os.Remove(absPath)
					if err != nil {
						// Ignore "not found" errors for delete, as it might already be gone
						if !os.IsNotExist(err) {
							errMsg := fmt.Sprintf("Failed to delete %s: %v", mod.Path, err)
							errorMessages = append(errorMessages, errMsg)
							errorCount++
						} else {
							// Optionally log that file was already deleted or treat as success
							fmt.Printf("File %s for deletion not found, already deleted or never existed.\n", mod.Path)
							appliedCount++ // Count as "applied" if goal is deletion and it's not there
						}
					} else {
						appliedCount++
					}
				default:
					errMsg := fmt.Sprintf("Unknown action '%s' for file %s", mod.Action, mod.Path)
					errorMessages = append(errorMessages, errMsg)
					errorCount++
				}
			}

			summaryMessage := fmt.Sprintf("Applied %d modifications successfully.", appliedCount)
			if errorCount > 0 {
				summaryMessage += fmt.Sprintf("\nEncountered %d errors:\n%s", errorCount, strings.Join(errorMessages, "\n"))
				dialog.ShowError(fmt.Errorf(summaryMessage), window) // Show as error if any errors occurred
			} else {
				dialog.ShowInformation("Apply Complete", summaryMessage, window)
			}
			statusBar.SetText(summaryMessage)

		}, window) // End of ShowConfirm
	})

	helpText := widget.NewLabel(
		"Paste the JSON response from the AI into the text area below.\n" +
			"The JSON should follow the format specified in the AI System Prompt (usually an object with a 'modified_files' array).\n" +
			"Each item in 'modified_files' should have 'path', 'content', and 'action' ('update', 'create', 'delete').\n" +
			"Paths must match those provided to the AI (e.g., 'project_root_basename/src/file.go').\n" +
			"**WARNING**: This operation will modify your local files. Ensure you have backups or use version control.",
	)
	helpText.Wrapping = fyne.TextWrapWord

	return container.NewVScroll(container.NewVBox(
		helpText,
		widget.NewSeparator(),
		aiResponseEntry,
		applyButton,
	))
}
