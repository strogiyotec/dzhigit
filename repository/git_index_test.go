package repository

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewIndex(t *testing.T) {
	//dzhigit repo dir
	dir := "/tmp/dzhigit" + Objects
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.RemoveAll(dir)
	//temp file to save in index
	file, err := ioutil.TempFile(dir, "tempFile")
	if err != nil {
		t.Fatal(err.Error())
	}
	content := "Some file content"
	file.WriteString(content)
	gitFile := DefaultGitFileFormatter{}
	objType, _ := AsGitObjectType("blob")
	//serialize it
	serialized, err := gitFile.Serialize([]byte(content), objType)
	if err != nil {
		t.Fatal(err.Error())
	}
	//save it
	err = gitFile.Save(serialized, dir)
	index, err := NewIndex(file.Name(), FILE, serialized.Hash, dir)
	if err != nil {
		t.Fatal(err.Error())
	}
	if index.hash != serialized.Hash {
		t.Fatalf("Wrong index hash , expected %s, got %s", serialized.Hash, index.hash)
	}
	if index.mode != FILE {
		t.Fatalf("Wrong index mode , expected %s, got %s", FILE, index.mode)
	}
	if index.path != file.Name() {
		t.Fatalf("Wrong index path , expected %s, got %s", file.Name(), index.path)
	}
}
