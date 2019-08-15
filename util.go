package cucumber

import (
	"os"
	"path"
)

const (
	featureFileExtension = ".feature"
)

func findFeatures(filepath string) ([]string, error) {
	var files []string

	fi, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		f, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		list, err := f.Readdir(-1)
		if err != nil {
			return nil, err
		}

		for _, fi := range list {
			if path.Ext(fi.Name()) == featureFileExtension {
				files = append(files, path.Join(filepath, fi.Name()))
			}
		}
	case mode.IsRegular():
		files = append(files, filepath)
	}

	return files, nil
}
