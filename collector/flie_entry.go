package collector

// FileEntry holds metadata for a file to be previewed or processed.
type FileEntry struct {
	Path         string `json:"path"`          // Relative path from root (including root's basename)
	SourcePath   string `json:"-"`             // Full system path to the file
	Mode         string `json:"-"`             // Collection mode ("path", "content", "both")
	Size         int64  `json:"size_bytes"`    // File size in bytes
	Format       string `json:"format"`        // File extension (e.g., "vue", "ts")
	OriginalPath string `json:"original_path"` // Relative path from root (as per os.Rel)
}

// ProcessedFile represents the data extracted from a file for the output JSON.
type ProcessedFile map[string]string // Typically {"path": "...", "content": "..."}
