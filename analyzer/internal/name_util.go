package internal

import "path/filepath"

func GetNameFromDirPath(path string) string {
	result := LabelName(filepath.Base(path))
	if result == "" {
		result = "root"
	}
	return result
}
