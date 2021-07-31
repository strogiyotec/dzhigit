package cli

import (
	"errors"
	"fmt"

	"github.com/strogiyotec/dzhigit/repository"
)

type GitAdd struct {
	files     []string
	reposPath string
}

//reader to read a file content by path
//use first class function to improve testability
type FileReader func(path string) ([]byte, error)

func GitCat(
	hash string,
	fileFormatter repository.GitFileFormatter,
	path string,
	reader FileReader,
) (*repository.DeserializedGitObject, error) {
	dir, fileName := repository.BlobDirWithFileName(hash)
	if !repository.Exists(path + dir + "/" + fileName) {
		return nil, errors.New(fmt.Sprintf("File with hash %s doesn't exist", hash))
	}
	data, err := reader(path + dir + "/" + fileName)
	if err != nil {
		return nil, err
	}
	return fileFormatter.Deserialize(data)
}

func NewGitAdd(files []string, repoPath string) (*GitAdd, error) {
	if repository.Exists(repoPath) {
		return &GitAdd{
			files:     files,
			reposPath: repoPath,
		}, nil
	} else {
		return nil, errors.New("Can't add git files, repository doesn't exist")
	}
}

//Add git files
func (command *GitAdd) Add() error {
	return nil
}
