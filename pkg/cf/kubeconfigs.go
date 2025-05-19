package cf

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type Candidate struct {
	Name     string
	FullPath string
}

func (c Candidate) Title() string {
	return c.Name
}

func (c Candidate) Description() string {
	return "Path: " + c.FullPath
}

func (c Candidate) FilterValue() string {
	return c.Name
}

// ListKubeconfigCandidatesInDir lists all files in dir that matches KubeconfigFilenamePattern
func ListKubeconfigCandidatesInDir(dir string) ([]Candidate, error) {
	fileInfo, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.Wrap(err, "os.ReadDir error")
	}

	var files []Candidate
	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		groupNames := kubeconfigFilenameMatchPattern.SubexpNames() // regex match groups
		nameGroupIndex := 0
		for i, name := range groupNames {
			if name == KubeconfigFilenameMatchPatternNameGroup { // find the "name" group index
				nameGroupIndex = i
				break
			}
		} // if there is no "name" group, will use the whole config file name

		matches := kubeconfigFilenameMatchPattern.FindStringSubmatch(file.Name())
		if len(matches) >= 2 {
			files = append(files, Candidate{
				// Use the last match group as the name, if there is no match group in the regex,
				// will use the whole config file name, I think this is the best we can do with different regex.
				Name:     matches[nameGroupIndex],
				FullPath: filepath.Join(dir, file.Name()),
			})
		}
	}
	return files, nil
}
