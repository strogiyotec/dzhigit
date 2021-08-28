package cli

import (
	"fmt"
	"testing"
	"time"

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

func Test_createCommitTree(t *testing.T) {
	treeHash, err := repository.GenerateHash([]byte("Tree Hash"))
	if err != nil {
		t.Fatal(err)
	}
	parentHash, err := repository.GenerateHash([]byte("Parent Hash"))
	if err != nil {
		t.Fatal(err)
	}
	user := &User{
		Name:  "Almas",
		Email: "almas337519@gmail.com",
	}
	time := &Time{
		zone:        "PDT",
		unixSeconds: time.Now().Unix(),
	}
	commit := Commit{
		treeHash:   repository.Hash(treeHash),
		message:    "New Commit",
		parentHash: repository.Hash(parentHash),
		user:       user,
		time:       time,
	}
	formatter := repository.DefaultGitFileFormatter{}
	commitObj, err := createCommitObject(commit, &formatter)
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

func Test_parseCommit(t *testing.T) {
	hash, err := repository.GenerateHash([]byte("Tree hash"))
	if err != nil {
		t.Fatal(err)
	}
	treeHash, err := repository.NewHash(hash)
	if err != nil {
		t.Fatal(err)
	}
	commitContent := fmt.Sprintf(
		`tree %s
		 parent %s
         author strogiyotec <almas337519@gmail.com> 1630023095 PDT
         comitter strogiyotec <almas337519@gmail.com> 1630023095 PDT

Message`,
		treeHash, treeHash)
	commit, err := parseCommit(commitContent)
	if err != nil {
		t.Fatal(err)
	}
	if commit.treeHash != treeHash {
		t.Fatalf(
			"Wrong tree hash, '%s' expected,got '%s'",
			treeHash,
			commit.treeHash,
		)
	}
	if commit.message != "Message" {
		t.Fatalf(
			"Wrong commit message, 'Message' expected,got '%s'",
			treeHash,
		)
	}
	if commit.user.Name != "strogiyotec" {

		t.Fatalf(
			"Wrong author name, 'strogiyotec' expected,got '%s'",
			commit.user.Name,
		)
	}
}

func TestUser_String(t *testing.T) {
	user := User{
		Name:  "strogiyotec",
		Email: "almas337519@gmail.com",
	}
	if user.String() != "strogiyotec almas337519@gmail.com" {
		t.Fatal("Wrong user's string representation")
	}
}
