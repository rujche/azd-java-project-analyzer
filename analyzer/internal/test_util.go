package internal

import (
	"os"
	"path/filepath"
)

type TestPom struct {
	PomFileRelativePath string
	PomContentString    string
}

func PrepareTestPomFiles(testPoms []TestPom) (string, error) {
	tempDir, err := os.MkdirTemp("", "PrepareTestPomFiles")
	if err != nil {
		return "", err
	}
	for _, testPom := range testPoms {
		pomPath := filepath.Join(tempDir, testPom.PomFileRelativePath)
		err := os.MkdirAll(filepath.Dir(pomPath), 0755)
		if err != nil {
			return "", err
		}
		err = os.WriteFile(pomPath, []byte(testPom.PomContentString), 0600)
		if err != nil {
			return "", err
		}
	}
	return tempDir, nil
}
