package appcfg

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/beclab/devbox/pkg/constants"
)

// UpdateMetadataField updates one field under metadata: by in-place line replacement to preserve templating and comments.
func UpdateMetadataField(appDir string, field string, value string) error {
	appCfgPath := filepath.Join(appDir, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgPath)
	if err != nil {
		return err
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	inMetadata := false
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if !inMetadata {
			if strings.TrimSpace(line) == "metadata:" {
				inMetadata = true
			}
			continue
		}
		if len(line) > 0 && (line[0] != ' ' && line[0] != '\t') {
			break
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, field+":") {
			colonIdx := strings.Index(line, ":")
			if colonIdx == -1 {
				continue
			}
			prefix := line[:colonIdx+1]
			after := line[colonIdx+1:]
			commentIdx := strings.Index(after, "#")
			var comment string
			if commentIdx >= 0 {
				comment = after[commentIdx:]
			}
			lines[i] = strings.TrimRight(prefix, " ") + " " + value
			if commentIdx >= 0 {
				lines[i] += " " + strings.TrimRight(comment, "\r\n")
			}
			break
		}
	}
	newContent := strings.Join(lines, "\n")
	return os.WriteFile(appCfgPath, []byte(newContent), 0644)
}

// UpdateEntrancesTitleAddDev appends "-dev" to every title under entrances list
// It performs in-place line updates to preserve existing comments and templating.
func UpdateEntrancesTitleAddDev(appDir string) error {
	appCfgPath := filepath.Join(appDir, constants.AppCfgFileName)
	data, err := os.ReadFile(appCfgPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	inEntrances := false
	entrancesIndent := ""

	// get leading whitespace (spaces/tabs) of a line
	leadingWS := func(s string) string {
		for i := 0; i < len(s); i++ {
			if s[i] != ' ' && s[i] != '\t' {
				return s[:i]
			}
		}
		return s
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if !inEntrances {
			if trimmed == "entrances:" {
				inEntrances = true
				entrancesIndent = leadingWS(line)
			}
			continue
		}

		// leave entrances block when indentation returns, but ignore helm template lines like {{- end }}
		if len(line) > 0 {
			ws := leadingWS(line)
			if len(ws) <= len(entrancesIndent) && strings.TrimSpace(line) != "" && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "{{") && !strings.HasPrefix(trimmed, "-") {
				inEntrances = false
				i-- // re-process this line outside entrances
				continue
			}
		}

		if strings.HasPrefix(trimmed, "title:") {
			colonIdx := strings.Index(line, ":")
			if colonIdx == -1 {
				continue
			}
			prefix := line[:colonIdx+1]
			after := line[colonIdx+1:]
			commentIdx := strings.Index(after, "#")
			var valuePart string
			var comment string
			if commentIdx >= 0 {
				valuePart = after[:commentIdx]
				comment = after[commentIdx:]
			} else {
				valuePart = after
			}

			v := strings.TrimSpace(valuePart)
			newV := v
			if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
				quote := v[0]
				content := v[1 : len(v)-1]
				if strings.HasSuffix(content, "-dev") {
					newV = v
				} else {
					newV = string(quote) + content + "-dev" + string(quote)
				}
			} else if v != "" {
				if strings.HasSuffix(v, "-dev") {
					newV = v
				} else {
					newV = v + "-dev"
				}
			} else {
				newV = "-dev"
			}

			newLine := strings.TrimRight(prefix, " ") + " " + newV
			if commentIdx >= 0 {
				newLine += " " + strings.TrimRight(comment, "\r\n")
			}
			lines[i] = newLine
		}
	}

	newContent := strings.Join(lines, "\n")
	return os.WriteFile(appCfgPath, []byte(newContent), 0644)
}
