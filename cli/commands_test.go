package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/strogiyotec/dzhigit/fakes"
	"github.com/strogiyotec/dzhigit/repository"
)

func TestNewUser(t *testing.T) {
	content := `{"name":"Almas","email":"almas337519@gmail.com"}`
	user, err := NewUser([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if user.Email != "almas337519@gmail.com" {
		t.Fatal("Wrong email for deserialized user")
	}
	if user.Name != "Almas" {
		t.Fatal("Wrong name for deserialized user")
	}
}

func TestWriteTree(t *testing.T) {
	dir, err := fakes.TempDir()
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

func Test_createCommitTree(t *testing.T) {
	treeHash, err := repository.GenerateHash([]byte("Tree Hash"))
	if err != nil {
		t.Fatal(err)
	}
	parentHash, err := repository.GenerateHash([]byte("Parent Hash"))
	if err != nil {
		t.Fatal(err)
	}
	commit := Commit{
		treeHash:   repository.Hash(treeHash),
		message:    "New Commit",
		parentHash: repository.Hash(parentHash),
	}
	user := User{
		Name:  "Almas",
		Email: "almas337519@gmail.com",
	}
	time := Time{
		zone:        "PDT",
		unixSeconds: time.Now().Unix(),
	}
	formatter := repository.DefaultGitFileFormatter{}
	commitObj, err := createCommitObject(commit, user, time, &formatter)
	if err != nil {
		t.Fatal(err)
	}
	deser, err := formatter.Deserialize(commitObj.Content)
	if err != nil {
		t.Fatal(err)
	}
	if deser.ObjType != repository.COMMIT {
		t.Fatalf(fmt.Sprintf("Wrong object type, 'commit expected', got %s", deser.ObjType))
	}
}
