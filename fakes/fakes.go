package fakes

import (
	"io/ioutil"
	"os"

	"github.com/strogiyotec/dzhigit/repository"
)

func FakeIndexEntries(entries []repository.SerializedGitObject, repo string) ([]repository.IndexEntry, error) {
	indexEntries := []repository.IndexEntry{}
	for _, entry := range entries {
		index, err := repository.NewIndex("first", repository.FILE, entry.Hash, repo)
		if err != nil {
			return nil, err
		}
		indexEntries = append(indexEntries, *index)
	}
	return indexEntries, nil
}

func TempDir() (string, error) {
	dir := "/tmp/dzhigit" + repository.Objects
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func FakeEntries(formatter repository.DefaultGitFileFormatter) ([]repository.SerializedGitObject, error) {
	firstEntry := []byte("First entry")
	secondEntry := []byte("Second entry")
	firstSer, err := formatter.Serialize(firstEntry, repository.BLOB)
	if err != nil {
		return nil, err
	}
	secSer, err := formatter.Serialize(secondEntry, repository.BLOB)
	if err != nil {
		return nil, err
	}
	return []repository.SerializedGitObject{
		*firstSer,
		*secSer,
	}, nil
}

func FakeFiles(dir string, amount int) ([]os.File, error) {
	files := []os.File{}
	for i := 0; i < amount; i++ {
		file, err := ioutil.TempFile(dir, "blob")
		if err != nil {
			return nil, err
		}
		files = append(files, *file)
	}
	return files, nil
}
