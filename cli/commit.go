package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/strogiyotec/dzhigit/repository"
	"github.com/tcnksm/go-gitconfig"
)

type Time struct {
	zone        string
	unixSeconds int64
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
type Commit struct {
	treeHash   repository.Hash //hash of a tree object
	message    string
	parentHash repository.Hash //hash of a parent commit may be null
	user       *User
	time       *Time
}

//Create a commit object
//This method does a validation and then delegates
//an actual commit creation to #createCommitTree
func CommitTree(
	commit Commit,
	path string,
	fileFormatter repository.GitFileFormatter,
	reader repository.FileReader,
) (*repository.SerializedGitObject, error) {
	tp, err := repository.TypeByHash(path, commit.treeHash, reader, fileFormatter)
	if err != nil {
		return nil, err
	}
	if tp != repository.TREE {
		return nil,
			errors.New(
				fmt.Sprintf(
					"Object with given hash '%s' is not a tree object",
					commit.treeHash,
				),
			)
	}
	if commit.HasParent() {
		tp, err := repository.TypeByHash(path, commit.parentHash, reader, fileFormatter)
		if err != nil {
			return nil, err
		}
		if tp != repository.COMMIT {
			return nil,
				errors.New(
					fmt.Sprintf(
						"Given hash %s is not a commit object",
						commit.parentHash,
					),
				)
		}
	}
	return createCommitObject(
		commit,
		fileFormatter,
	)
}
func (c *Commit) HasParent() bool {
	return len(c.parentHash) != 0
}

func (u *User) String() string {
	return fmt.Sprintf("%s %s", u.Name, u.Email)
}

func (t *Time) String() string {
	commitTime := time.Unix(t.unixSeconds, 0)
	//TODO: it doesn't use timezone yet
	return commitTime.Format(time.RFC3339)
}

func NewCommit(
	hash repository.Hash,
	message string,
	parent repository.Hash,
	user *User,
	time *Time,
) *Commit {
	return &Commit{
		treeHash:   hash,
		message:    message,
		parentHash: parent,
		user:       user,
		time:       time,
	}
}
func CurrentTime() *Time {
	now := time.Now()
	zone, _ := now.Zone()
	seconds := now.Unix()
	return &Time{
		zone:        zone,
		unixSeconds: seconds,
	}
}

func NewUser(content []byte) (*User, error) {
	var user User
	err := json.Unmarshal(content, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func DefaultGitUserAsJson() ([]byte, error) {
	username, err := gitconfig.Username()
	if err != nil {
		return nil, err
	}
	email, err := gitconfig.Email()
	if err != nil {
		return nil, err
	}
	return json.Marshal(
		User{
			Name:  username,
			Email: email,
		},
	)
}

// +----------------------------+
// | Commit format line by line |
// +----------------------------+
// | tree hash                  |
// | parent hash                |
// | author                     |
// | comitter                   |
// | empty line                 |
// | commit message             |
// +----------------------------+
func createCommitObject(
	commit Commit,
	formatter repository.GitFileFormatter,
) (*repository.SerializedGitObject, error) {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("tree %s\n", commit.treeHash))
	if commit.HasParent() {
		builder.WriteString(fmt.Sprintf("parent %s\n", commit.parentHash))
	}
	builder.WriteString(
		fmt.Sprintf(
			"author %s <%s> %d %s\n",
			commit.user.Name,
			commit.user.Email,
			commit.time.unixSeconds,
			commit.time.zone,
		),
	)
	builder.WriteString(
		fmt.Sprintf(
			"comitter %s <%s> %d %s\n",
			commit.user.Name,
			commit.user.Email,
			commit.time.unixSeconds,
			commit.time.zone,
		),
	)
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("%s\n", commit.message))
	return formatter.Serialize([]byte(builder.String()), repository.COMMIT)
}

func parseCommit(content string) (*Commit, error) {
	parts := strings.Split(content, "\n")
	commit := &Commit{}
	treeHash, err := repository.NewHash(strings.Fields(parts[0])[1])
	if err != nil {
		return nil, err
	}
	commit.treeHash = treeHash
	nextIndex := 1
	if strings.Contains(parts[1], "parent") {
		parentHash, err := repository.NewHash(strings.Fields(parts[1])[1])
		if err != nil {
			return nil, err
		}
		commit.parentHash = parentHash
		nextIndex++
	}
	authorParts := strings.Fields(parts[nextIndex])
	commit.user = &User{
		Name:  authorParts[1],
		Email: authorParts[2],
	}
	timeUnix, err := strconv.ParseInt(authorParts[3], 10, 64)
	if err != nil {
		return nil, err
	}
	commit.time = &Time{
		unixSeconds: timeUnix,
		zone:        authorParts[4],
	}
	//skip committer and empty line
	nextIndex += 3
	commit.message = strings.Join(parts[nextIndex:], "\n")
	return commit, nil
}
