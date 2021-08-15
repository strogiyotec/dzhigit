package cli

import (
	"os"
	"testing"

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
	entries, err := fakes.FakeEntries(formatter)
	if err != nil {
		t.Fatal(err)
	}
	files, err := fakes.FakeFiles(dir, len(entries))
	if err != nil {
		t.Fatal(err)
	}
}
