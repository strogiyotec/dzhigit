package repository

import (
	"errors"
	"fmt"
	"syscall"
)

type Mode string

const (
	FILE       Mode = "100644"
	EXECUTABLE      = "100755"
)

func AsMode(mode string) (Mode, error) {
	switch mode {
	case "100644":
		return FILE, nil
	case "100755":
		return EXECUTABLE, nil
	default:
		return "", errors.New("Invalid file mode")
	}
}

type IndexEntry struct {
	path             string
	mode             Mode
	creationTime     int64
	modificationTime int64
	hash             string
}

//TODO: add test
func NewIndex(file, plainMode, hash, repoPath string) (*IndexEntry, error) {
	if !Exists(file) {
		return nil, errors.New(fmt.Sprintf("File %s doesn't exist", file))
	}
	mode, err := AsMode(plainMode)
	if err != nil {
		return nil, err
	}
	dir, fileName := BlobDirWithFileName(hash)
	if !Exists(repoPath + dir + "/" + fileName) {
		return nil, errors.New(fmt.Sprintf("File with hash %s doesn't exist", hash))
	}
	modeTime, crTime, err := getTimes(file)
	if err != nil {
		return nil, err
	}
	return &IndexEntry{
		path:             file,
		mode:             mode,
		creationTime:     crTime,
		modificationTime: modeTime,
		hash:             hash,
	}, nil
}

func getTimes(path string) (int64, int64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return -1, -1, err
	}
	return st.Mtim.Sec, st.Ctim.Sec, nil
}

//TODO: add test
func (index IndexEntry) String() string {
	//Mode C_time M_time sha1-hash F_name
	return fmt.Sprintf(
		"%s %d %d %s\t%s",
		index.mode,
		index.creationTime,
		index.modificationTime,
		index.hash,
		index.path,
	)
}

type GitIndex interface {
	Add(IndexEntry) error
}