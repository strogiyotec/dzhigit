package repository

import "errors"

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
	changeTime       int64
	modificationTime int64
	hash             string
}

type GitIndex interface {
	Add(IndexEntry) error
}
