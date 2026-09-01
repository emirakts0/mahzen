package searchutil

// ContentExcerptLen is the maximum length of content excerpts returned in search results.
const ContentExcerptLen = 300

// textReadableTypes is the set of file extensions whose content is
// human-readable and should be indexed for full-text search.
var textReadableTypes = map[string]struct{}{
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

// IsTextReadable reports whether content for the file extension should be indexed.
func IsTextReadable(fileType string) bool {
	_, ok := textReadableTypes[fileType]
	return ok
}
