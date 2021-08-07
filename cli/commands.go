package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/strogiyotec/dzhigit/repository"
)

type GitAdd struct {
	files     []string
	reposPath string
}

//reader to read a file content by path
//use first class function to improve testability
type FileReader func(path string) ([]byte, error)

func WriteTree(indexLines []string) ([]byte, error) {
	indexes := []repository.IndexEntry{}
	for _, line := range indexLines {
		index, err := repository.ParseLineToIndex(line)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, *index)
	}
	//map key is a file level and value is a list of indexes in this level
	levels := make(map[int][]repository.IndexEntry)
	for _, index := range indexes {
		depth := index.Depth()
		if val, ok := levels[depth]; ok {
			val = append(val, index)
		} else {
			val := []repository.IndexEntry{}
			val = append(val, index)
			levels[depth] = val
		}
	}
	fmt.Println(levels)
	return []byte(string("test")), nil
}

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

func UpdateIndex(index repository.IndexEntry, path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return repository.Add(index, f)
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
