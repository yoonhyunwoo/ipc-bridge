//go:build windows

package endpoint

import "os"

func parseSocketStat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
