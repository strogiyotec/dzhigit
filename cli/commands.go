package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/strogiyotec/dzhigit/repository"
)

type GitAdd struct {
	files     []string
	reposPath string
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewUser(content []byte) (*User, error) {
	var user User
	err := json.Unmarshal(content, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

//reader to read a file content by path
//use first class function to improve testability
type FileReader func(path string) ([]byte, error)

func CommitTree(
	message string,
	hash string,
	currentTime int64,
	timeZoneName string,
	path string,
	user User,
	gitFileSystem repository.GitFileFormatter,
) (*repository.SerializedGitObject, error) {
	dir, fileName := repository.BlobDirWithFileName(hash)
	if !repository.Exists(path + dir + "/" + fileName) {
		return nil,
			errors.New(
				fmt.Sprintf(
					"File with hash %s doesn't exist",
					hash,
				),
			)
	}
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("tree %s\n", hash))
	builder.WriteString(
		fmt.Sprintf(
			"author %s <%s> %d %s\n",
			user.Name,
			user.Email,
			currentTime,
			timeZoneName,
		),
	)
	builder.WriteString(
		fmt.Sprintf(
			"comitter %s <%s> %d %s\n",
			user.Name,
			user.Email,
			currentTime,
			timeZoneName,
		),
	)
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("%s\n", message))
	return gitFileSystem.Serialize([]byte(builder.String()), repository.COMMIT)
}

//TODO: write a test
func createTreeEntry(
	level int,
	indexes []repository.IndexEntry,
	fileFormatter repository.GitFileFormatter,
	objPath string,
) (*repository.SerializedGitObject, error) {
	if len(indexes) == 0 {
		return nil, nil
	}
	nextLevels := make(map[string][]repository.IndexEntry)
	builder := strings.Builder{}
	for _, index := range indexes {
		if index.Depth() == level {
			builder.WriteString(index.BlobString(level - 1))
		} else {
			if val, ok := nextLevels[index.PathParts()[level-1]]; ok {
				val = append(val, index)
				nextLevels[index.PathParts()[level-1]] = val
			} else {
				var entries []repository.IndexEntry
				entries = append(entries, index)
				nextLevels[index.PathParts()[level-1]] = entries
			}
		}
	}
	for key, elements := range nextLevels {
		tree, err := createTreeEntry(level+1, elements, fileFormatter, objPath)
		if err != nil {
			return nil, err
		}
		if tree != nil {
			builder.WriteString(treeLine(tree.Hash, key))
		}
	}
	tree, err := fileFormatter.Serialize(
		[]byte(builder.String()),
		repository.TREE,
	)
	if err != nil {
		return nil, err
	}
	err = fileFormatter.Save(tree, objPath)
	if err != nil {
		//if tree already exists then just return it
		if _, ok := err.(*repository.TreeAlreadyExistError); ok {
			return tree, nil
		}
		return nil, err
	}
	return tree, nil
}

//string that shows how a reference to a tree is stored inside of a tree file
//TODO: can we unite it with Index Entry Blob String ?
func treeLine(hash, dir string) string {
	return fmt.Sprintf("040000 tree %s\t%s\n", hash, dir)
}

func WriteTree(
	indexLines []string,
	objPath string,
	gitFormatter repository.GitFileFormatter,
) (*repository.SerializedGitObject, error) {
	var indexes []repository.IndexEntry
	for _, line := range indexLines {
		index, err := repository.ParseLineToIndex(line)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, *index)
	}
	return createTreeEntry(1, indexes, gitFormatter, objPath)
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
		return nil, errors.New("Can't add git files, repository doesn't exist ")
	}
}

//Add git files
func (command *GitAdd) Add() error {
	return nil
}
