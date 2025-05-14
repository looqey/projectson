package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const configDocsMarkdown = `
# projectson configuration

This document describes the fields available in the ` + "`yaml`" + ` file used by projectson.

---

## ` + "`root`" + `
-   **Type**: ` + "`String`" + `
-   **Required**: Yes
-   **Description**: The absolute path to the root directory of the project you want to analyze.
-   **Example**:
` + "```yaml" + `
root: "/path/to/your/project"
# On Windows:
# root: "C:\\Users\\YourName\\Projects\\MyProject"
` + "```" + `

---

## ` + "`include`" + `
-   **Type**: ` + "`List of Strings`" + `
-   **Required**: No (Defaults to scanning the entire ` + "`root`" + ` directory, respecting ` + "`formats`" + ` and ` + "`exclude_patterns`" + `)
-   **Description**: A list of specific files or directories to include for processing. Paths are relative to the ` + "`root`" + ` directory. Each entry can define a path and an optional collection mode.
-   **Syntax per entry**:
    -   ` + "`\"path/to/item\"`" + `: Collects both path and content (default mode). If ` + "`item`" + ` is a directory, it's scanned recursively.
    -   ` + "`\"path/to/item:path\"`" + `: Collects only the path string for the item(s).
    -   ` + "`\"path/to/item:content\"`" + `: Collects only the content for the item(s).
    -   ` + "`\"path/to/item:both\"`" + `: Explicitly collects both path and content.
    -   ` + "`\"path/to/directory/*\"`" + `: Collects files *directly within* ` + "`path/to/directory`" + ` (non-recursively). Default mode is ` + "`both`" + `.
    -   ` + "`\"path/to/directory/*:mode\"`" + `: Collects files *directly within* ` + "`path/to/directory`" + ` with the specified ` + "`mode`" + `.
-   **Examples**:
` + "```yaml" + `
include:
  - "src"  # Include everything in src/, recursively, path & content
  - "README.md:content" # Include only content of README.md
  - "assets/*:path"     # Include only paths of files directly in assets/
  - "docs/api.md"       # Include path & content of docs/api.md
` + "```" + `

---

## ` + "`formats`" + `
-   **Type**: ` + "`List of Strings`" + `
-   **Required**: Yes
-   **Description**: A list of file extensions (without the leading dot) to be included in the collection. Only files matching these extensions will be considered.
-   **Example**:
` + "```yaml" + `
formats:
  - "go"
  - "html"
  - "css"
  - "js"
  - "vue"
  - "ts"
  - "py"
` + "```" + `

---

## ` + "`output`" + `
-   **Type**: ` + "`String`" + `
-   **Required**: Yes
-   **Description**: The full path (or path relative to where ProjectSon is run) for the output JSON file where the collected data will be saved. If the directory does not exist, ProjectSon will attempt to create it.
-   **Example**:
` + "```yaml" + `
output: "project_data_output.json"
# output: "/tmp/my_project_collection.json"
` + "```" + `

---

## ` + "`exclude_patterns`" + `
-   **Type**: ` + "`List of Strings`" + `
-   **Required**: No
-   **Description**: A list of patterns to exclude files or directories. These patterns are applied *after* ` + "`include`" + ` rules and ` + "`formats`" + ` have identified potential candidates.
    -   **Glob Patterns**: Standard file globbing (e.g., ` + "`node_modules`" + `, ` + "`*.log`" + `, ` + "`dist/*`" + `). These match against:
        1.  The base name of the file or directory (e.g., ` + "`*.log`" + ` matches ` + "`debug.log`" + `).
        2.  The path relative to ` + "`root`" + ` (e.g., ` + "`src/tests/*`" + ` matches ` + "`src/tests/test_helper.go`" + ` if the pattern contains a path separator).
    -   **Regular Expressions**: Go-compatible regular expressions. Must be enclosed in forward slashes (e.g., ` + "`/\\.git/`" + `, ` + "`/private_.*\\.key$/`" + `). These also match against basenames or relative paths.
-   **Note**: If a directory is excluded, its contents will not be scanned.
-   **Examples**:
` + "```yaml" + `
exclude_patterns:
  - "node_modules"  # Exclude the entire node_modules directory
  - ".git"          # Exclude .git directory
  - "*.min.js"      # Exclude minified JavaScript files
  - "dist"          # Exclude build output directory
  - "/test_data/"   # Exclude directories named test_data (regex matching the name)
  - "/^\\.(svn|hg|DS_Store)/" # Exclude common VCS and OS files (regex)
  - "target/*"      # Exclude contents of a target directory (glob)
` + "```" + `

---

## ` + "`content_exclusions`" + `
-   **Type**: ` + "`List of Objects`" + `
-   **Required**: No
-   **Description**: A list of rules to remove specific sections of content from files *before* they are added to the output. This is useful for removing comments, boilerplate, sensitive data, or irrelevant code blocks. Each rule object has the following fields:
    -   ` + "`type` (String, Required)" + `: Specifies the type of exclusion.
        -   ` + "`\"delimiters\"`" + `: Uses start and end string tags to identify content to remove.
        -   ` + "`\"regexp\"`" + `: Uses a regular expression to identify content to remove.
    -   ` + "`file_pattern` (String, Required)" + `: A glob pattern that specifies which files this rule applies to, based on their extension.
        -   It is tested against the file's extension string (e.g., ` + "`file_pattern: \"vue\"`" + ` would match files with a ` + "`.vue`" + ` extension, as the code tests against ` + "`\"vue\"`" + `).
        -   It is also tested against the file's extension string prefixed with a dot (e.g., ` + "`file_pattern: \"*.vue\"`" + ` would match files with a ` + "`.vue`" + ` extension, as the code tests against ` + "`\".vue\"`" + `).
        -   Use ` + "`\"*\"`" + ` to apply the rule to all files that match the global ` + "`formats`" + ` list.
        -   **Note**: This pattern primarily matches based on the file's extension part, not the full filename directly (unless the extension itself forms the unique part of the filename you wish to target).
    -   ` + "`start` (String, Optional)" + `: Used when ` + "`type`" + ` is ` + "`\"delimiters\"`" + `. The starting string tag of the content to exclude.
    -   ` + "`end` (String, Optional)" + `: Used when ` + "`type`" + ` is ` + "`\"delimiters\"`" + `. The ending string tag of the content to exclude.
    -   ` + "`pattern` (String, Optional)" + `: Used when ` + "`type`" + ` is ` + "`\"regexp\"`" + `. The Go-compatible regular expression. The regex should be crafted to match the content you want to remove. It's often useful to use the ` + "`(?s)`" + ` flag (dot matches newline) for multi-line patterns.
-   **Examples**:
` + "```yaml" + `
content_exclusions:
  - type: "delimiters"
    file_pattern: "*.vue" # Or "vue"
    start: "<style>"
    end: "</style>"
  - type: "delimiters"
    file_pattern: "html"
    start: "<!-- BEGIN_EXCLUDE -->"
    end: "<!-- END_EXCLUDE -->"
  - type: "regexp"
    file_pattern: "*.js" # Or "js"
    pattern: "(//.*)|(/\\*[\\s\\S]*?\\*/)" # Remove JS line and block comments
  - type: "regexp"
    file_pattern: "*" # Apply to all matched file types
    pattern: "SECRET_API_KEY = '.*?'" # Remove a line containing a secret key
  - type: "regexp"
    file_pattern: "java"
    # Example: Remove a specific Java annotation block or a generated class/method.
    # This regex might need adjustments based on specific code style.
    pattern: "@Generated(?:\\s*\\([^)]*\\))?\\s*(?:public\\s+|protected\\s+|private\\s+)?(?:static\\s+|final\\s+)?(?:class|interface|enum|@interface|\\S+\\s+\\S+\\s*\\([^)]*\\))\\s*\\S+\\s*(?:\\{[\\s\\S]*?\\}|;)"
` + "```" + `
`

// MakeConfigDocsPage creates the UI for displaying the configuration documentation.
func MakeConfigDocsPage() fyne.CanvasObject {
	richText := widget.NewRichTextFromMarkdown(configDocsMarkdown)
	richText.Wrapping = fyne.TextWrapWord

	return container.NewVScroll(richText)
}
