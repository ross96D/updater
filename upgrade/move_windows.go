//go:build windows

package upgrade

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// TODO maybe implement the fix of path to be larger than 256 characters
func move(src, dst string) error {
	temp := dst + ".old"
	err := moveFileEx(dst, temp)
	if err != nil {
		return err
	}
	return moveFileEx(src, dst)
}

func moveFileEx(oldPath, newPath string) error {
	lpOldPath, err := windows.UTF16PtrFromString(oldPath)
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString moveFileEx %s %w", oldPath, err)
	}
	lpNewPath, err := windows.UTF16PtrFromString(newPath)
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString moveFileEx %s %w", newPath, err)
	}
	err = windows.MoveFileEx(lpOldPath, lpNewPath, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_COPY_ALLOWED)
	if err != nil {
		return fmt.Errorf("moveFileEx %s to %s %w", oldPath, newPath, err)
	}
	return nil
}

func deleteOnReboot(path string) error {
	lpPath, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return fmt.Errorf("UTF16PtrFromString moveFileEx %s %w", path, err)
	}
	err = windows.MoveFileEx(lpPath, nil, windows.MOVEFILE_DELAY_UNTIL_REBOOT)
	if err != nil {
		return fmt.Errorf("moveFileEx %s to NULL %w", path, err)
	}
	return nil
}
