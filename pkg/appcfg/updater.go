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


