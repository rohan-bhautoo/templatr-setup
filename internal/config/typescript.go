package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// UpdateConfigFile reads a TypeScript/JavaScript config file and replaces
// values for the given field paths. Uses regex-based pattern matching -
// intentionally not a full AST parser.
//
// Paths like "siteConfig.name" are matched by finding the key "name"
// followed by a string value in the file. The last path component is
// used as the key to match.
func UpdateConfigFile(path string, fieldValues map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", path, err)
	}

	content := string(data)

	for fieldPath, newValue := range fieldValues {
		// Extract the key name (last component of the path)
		parts := strings.Split(fieldPath, ".")
		key := parts[len(parts)-1]

		content = replaceFieldValue(content, key, newValue)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

// replaceFieldValue finds a key-value pattern in TypeScript/JavaScript
// and replaces the string value.
//
// Matches patterns like:
//
//	name: "old value",
//	name: 'old value',
//	name: "old value"   // with or without trailing comma
func replaceFieldValue(content, key, newValue string) string {
	// Pattern: key followed by colon, optional whitespace, then a quoted string
	// Captures: the full match so we can replace just the value part
	pattern := fmt.Sprintf(`(\b%s\s*:\s*)("(?:[^"\\]|\\.)*"|'(?:[^'\\]|\\.)*')`, regexp.QuoteMeta(key))
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Find where the value starts
		loc := re.FindStringSubmatchIndex(match)
		if loc == nil {
			return match
		}

		// loc[2]:loc[3] is the key+colon prefix
		prefix := match[loc[2]:loc[3]]

		// Determine quote style from original
		origValue := match[loc[4]:loc[5]]
		quote := string(origValue[0]) // " or '

		return prefix + quote + escapeJSString(newValue, quote) + quote
	})
}

// escapeJSString escapes a string for use in a JavaScript string literal.
func escapeJSString(s, quote string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, quote, `\`+quote)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
