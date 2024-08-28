//go:build linux

package upgrade

import "os"

func move(src, dst string) error {
	temp := dst + ".old"

	err := os.Rename(dst, temp)
	if err != nil {
		return err
	}

	return os.Rename(src, dst)
}
