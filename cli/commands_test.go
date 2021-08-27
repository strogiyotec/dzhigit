package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/strogiyotec/dzhigit/fakes"
	"github.com/strogiyotec/dzhigit/repository"
)

func TestWriteTree(t *testing.T) {
	user, err := DefaultGitUserAsJson()
	if err != nil {
		t.Fatal(err)
	}
	dir, err := fakes.TempDir(user)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	formatter := repository.DefaultGitFileFormatter{}
	gitDir := dir + "/.dzhigit"
	//Create and save a bunch of blobs
	entries, err := fakes.FakeEntries(formatter)
	for _, entry := range entries {
		err = formatter.Save(&entry, repository.ObjPath(gitDir))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	files, err := fakes.FakeFiles(dir, len(entries))
	if err != nil {
		t.Fatal(err)
	}
	//Save these blobs in index file
	objPath := repository.ObjPath(gitDir)
	indexEntries, err := fakes.FakeIndexEntries(entries, files, objPath)
	if err != nil {
		t.Fatal(err)
	}
	indexPath := repository.IndexPath(gitDir)
	for _, entry := range indexEntries {
		err = UpdateIndex(entry, indexPath)
		if err != nil {
			t.Fatal(err)
		}
	}
	//Create a tree from this index
	indexContent, err := ioutil.ReadFile(indexPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	lines := strings.Split(strings.TrimSpace(string(indexContent)), "\n")
	tree, err := WriteTree(lines, objPath, &formatter)
	if err != nil {
		t.Fatal(err)
	}
	objType, err := repository.TypeByHash(objPath, tree.Hash, repository.Reader, &formatter)
	if err != nil {
		t.Fatal(err)
	}
	if objType != repository.TREE {
		t.Fatalf(fmt.Sprintf("Wrong object type, 'tree' expected, got %s", objType))
	}

}
