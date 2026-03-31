package searchutil

// ContentExcerptLen is the maximum length of content excerpts returned in search results.
const ContentExcerptLen = 300

// TextReadableTypes is the set of file extensions that contain human-readable text
// and whose content should be indexed for full-text search.
var TextReadableTypes = map[string]struct{}{
	"":      {}, // plain text entry (no file type)
	"txt":   {},
	"md":    {},
	"go":    {},
	"java":  {},
	"py":    {},
	"js":    {},
	"ts":    {},
	"jsx":   {},
	"tsx":   {},
	"css":   {},
	"html":  {},
	"htm":   {},
	"xml":   {},
	"json":  {},
	"yaml":  {},
	"yml":   {},
	"toml":  {},
	"ini":   {},
	"sh":    {},
	"bash":  {},
	"zsh":   {},
	"rs":    {},
	"c":     {},
	"cpp":   {},
	"h":     {},
	"hpp":   {},
	"cs":    {},
	"rb":    {},
	"php":   {},
	"sql":   {},
	"r":     {},
	"kt":    {},
	"swift": {},
	"scala": {},
	"lua":   {},
	"pl":    {},
	"csv":   {},
	"log":   {},
	"conf":  {},
	"env":   {},
	"tf":    {},
}

// IsTextReadable reports whether the given file extension corresponds to a
// human-readable text format that should have its content indexed.
func IsTextReadable(fileType string) bool {
	_, ok := TextReadableTypes[fileType]
	return ok
}
