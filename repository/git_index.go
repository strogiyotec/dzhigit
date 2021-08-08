package repository

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Mode string

const (
	FILE       Mode = "100644"
	EXECUTABLE      = "100755"
)

const INDEX_PARTS = 5

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

func Add(entry IndexEntry, writer io.Writer) error {
	_, err := writer.Write([]byte(entry.String() + "\n"))
	return err
}

type IndexEntry struct {
	path             string
	mode             Mode
	creationTime     int64
	modificationTime int64
	hash             string
}

func (entry IndexEntry) PathParts() []string {
	return strings.Split(entry.path, string(os.PathSeparator))
}

//Get the depth of a file for given index
func (entry IndexEntry) Depth() int {
	parts := strings.Split(entry.path, string(os.PathSeparator))
	return len(parts)
}

//file - the file's name to index
//plainMode - the files' type
//hash - the hash of this file to index
//repoPath - path to repository
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

//Parse given line to index entry
func ParseLineToIndex(line string) (*IndexEntry, error) {
	parts := strings.Fields(line)
	if len(parts) != INDEX_PARTS {
		return nil,
			errors.New(
				fmt.Sprintf(
					"Invalid line for index, should containt %d parts, was %d",
					INDEX_PARTS,
					len(parts),
				),
			)
	}
	mode, err := AsMode(parts[0])
	if err != nil {
		return nil, errors.New("Unknown index type " + parts[0])
	}
	crTime, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid creation time, long expected")
	}
	modTime, err := strconv.ParseInt(parts[2], 10, 64)
	hash := parts[3]
	path := parts[4]
	return &IndexEntry{
		mode:             mode,
		creationTime:     crTime,
		modificationTime: modTime,
		hash:             hash,
		path:             path,
	}, nil
}

func getTimes(path string) (int64, int64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return -1, -1, err
	}
	return st.Mtim.Sec, st.Ctim.Sec, nil
}

func (index IndexEntry) TreeString() string {
	return fmt.Sprintf(
		"%s %s blob\t%s",
		index.mode,
		index.hash,
		index.path,
	)
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
