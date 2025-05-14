package collector

// AIFileModification represents a single file modification suggested by the AI.
type AIFileModification struct {
	Path    string `json:"path"`    // Relative path from project root (e.g., "src/main.go") as provided in the input.
	Content string `json:"content"` // Full new content of the file.
	Action  string `json:"action"`  // "update", "create", or "delete".
}

// AIResponse is the expected structure of the JSON response from the AI
// when it suggests file modifications.
type AIResponse struct {
	ModifiedFiles []AIFileModification `json:"modified_files"`
}
