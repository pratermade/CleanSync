package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

// WalkAndHash walks the directory structure that is specifed in the Syncer.Folderpath.
// Will filter for filetypes listed in the filters slice.
// Returns a map of filepath[lastModDate]
func WalkAndHash(filters []string, folderPath string) (map[string]int64, error) {

	retMap := make(map[string]int64)
	err := filepath.Walk(folderPath, func(p string, info os.FileInfo, err error) error {
		if err == nil {
			if !info.IsDir() {
				if !inFilters(info.Name(), filters) {
					return nil
				}
				h, err := getLastModDate(p)
				if err != nil {
					return err
				}
				p := Localize(p)
				retMap[p] = h
			}

		}
		return nil
	})
	if err != nil {
		// spinnerInfo.Fail(err)
		return nil, err
	}
	// spinnerInfo.Success("Taking Inventory of local files.")
	return retMap, nil
}

// localize converts paths to windows paths if needed, has its own function for future needs.
func Localize(s string) string {
	s = filepath.FromSlash(s)
	return s

}

// inFilters checks to see if the name of the file has one of the extensions listed in the filters slice, it returns true.
func inFilters(name string, filters []string) bool {
	for _, filter := range filters {
		if strings.HasSuffix(name, filter) {
			return true
		}
	}
	return false
}

// get lastModDate returns the last moidified date for the file specified by f (file path).
// Returns unix time
func getLastModDate(f string) (int64, error) {
	fileinfo, err := os.Stat(f)
	if err != nil {
		return 0, err
	}
	atime := fileinfo.ModTime().Unix()
	return atime, nil
}
