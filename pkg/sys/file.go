package sys

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func IsSymlink(stat os.FileInfo) bool {
	return stat.Mode()&os.ModeSymlink != 0
}

// GenerateBackUpName generates a backup name for a file,
// it will try to find a name that is not exist,
// if there are too many backup revisions, it will return an error.
func GenerateBackUpName(basePath, suffix string) (string, error) {
	index := 0
	var backupPath string
	for range 999 {
		if index == 0 {
			backupPath = fmt.Sprintf("%s%s", basePath, suffix)
		} else {
			backupPath = fmt.Sprintf("%s-%d%s", basePath, index, suffix)
		}
		_, err := os.Lstat(backupPath)
		if err == nil {
			index++ // backupPath already exists, try bigger index
			continue
		}

		// unknown error
		if !os.IsNotExist(err) {
			return "", errors.Wrap(err, "os.Lstat error")
		}

		// backupPath not exist
		return backupPath, nil
	}

	return "", errors.New("Too many backup revisions of this file")
}

// BackUpFile backs up a file, back up file will be named like filePath-backup-n
func BackUpFile(filePath string) error {
	backupPath, err := GenerateBackUpName(filePath+"-backup", "")
	if err != nil {
		return err
	}
	if err := os.Rename(filePath, backupPath); err != nil {
		return errors.Wrap(err, "os.Rename error")
	}
	return nil
}

// CreateSymlink creates newname as a symbolic link to oldname,
// if newname not exist or is a symlink, it will be replaced directly,
// in other cases, it will be backed up first
func CreateSymlink(linkToOldName, linkFromNewName string) error {
	newStat, err := os.Lstat(linkFromNewName)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Symlink(linkToOldName, linkFromNewName); err != nil {
				return errors.Wrap(err, "create new symlink error")
			}
			return nil
		}
		return errors.Wrap(err, "os.Lstat error")
	}

	if IsSymlink(newStat) {
		// is a symlink
		if err := os.Remove(linkFromNewName); err != nil {
			return errors.Wrap(err, "remove old symlink error")
		}
	} else {
		// is not a symlink
		if err := BackUpFile(linkFromNewName); err != nil {
			return errors.Wrap(err, "back up error")
		}
	}

	if err := os.Symlink(linkToOldName, linkFromNewName); err != nil {
		return errors.Wrap(err, "create symlink error")
	}
	return nil
}
