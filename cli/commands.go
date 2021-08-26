package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/strogiyotec/dzhigit/repository"
	"github.com/tcnksm/go-gitconfig"
)

type Time struct {
	zone        string
	unixSeconds int64
}

//how single entry in tree object is represented
type treeEntry struct {
	mode    repository.Mode
	objType repository.GitObjectType
	path    string
	hash    repository.Hash
}

//tuple for recursive checkout
//that contains a hash of a tree and a path that this tree represents
type checkoutTuple struct {
	treeHash repository.Hash
	path     string
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

func (c *Commit) HasParent() bool {
	return len(c.parentHash) != 0
}

func (u *User) String() string {
	return fmt.Sprintf("%s %s", u.Name, u.Email)
}

func (t *Time) String() string {
	commitTime := time.Unix(t.unixSeconds, 0)
	return commitTime.Format(time.RFC3339)
}

func Log(
	writer *tablewriter.Table,
	gitRepoPath string,
	formatter repository.GitFileFormatter,
	reader repository.FileReader,
) error {
	headPath := repository.HeadPath(gitRepoPath)
	content, err := os.ReadFile(headPath)
	if err != nil {
		return errors.New(
			`There is no branch in this repo,
            to create one use 'dzhigit update-ref'
            and then check it out using 'dzhigit checkout' `,
		)
	}
	branch := branchNameFromHead(string(content))
	headsPath := repository.HeadsPath(gitRepoPath)
	pathToBranch := headsPath + branch
	if !repository.Exists(pathToBranch) {
		return errors.New(
			fmt.Sprintf(
				"error branch with name '%s' doesn't exist",
				branch,
			),
		)
	}
	commitHashContent, err := reader(pathToBranch)
	if err != nil {
		return err
	}
	commitHash, err := repository.NewHash(string(commitHashContent))
	if err != nil {
		return err
	}
	objPath := repository.ObjPath(gitRepoPath)
	return appendLog(
		writer,
		commitHash,
		objPath,
		formatter,
		reader,
	)
}

func appendLog(
	writer *tablewriter.Table,
	commitHash repository.Hash,
	objPath string,
	formatter repository.GitFileFormatter,
	reader repository.FileReader,
) error {
	rawContent, err := reader(commitHash.Path(objPath))
	if err != nil {
		return err
	}
	deser, err := formatter.Deserialize(rawContent)
	if err != nil {
		return err
	}
	commit, err := parseCommit(deser.Content)
	if err != nil {
		return err
	}
	writer.Append(
		[]string{
			string(commit.treeHash)[0:5],
			commit.message,
			commit.user.String(),
			commit.time.String(),
		},
	)
	if commit.HasParent() {
		return appendLog(
			writer,
			commit.parentHash,
			objPath,
			formatter,
			reader,
		)
	}
	return nil
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
	nextIndex += 2
	commit.message = strings.Join(parts[nextIndex:], "\n")
	return commit, nil
}

func newTreeEntry(line string) (*treeEntry, error) {
	parts := strings.Split(line, "\\s+")
	mode, err := repository.AsMode(parts[0])
	if err != nil {
		return nil, err
	}
	objType, err := repository.AsGitObjectType(parts[1])
	if err != nil {
		return nil, err
	}
	hash, err := repository.NewHash(parts[2])
	if err != nil {
		return nil, err
	}
	path := parts[3]
	return &treeEntry{
		mode:    mode,
		objType: objType,
		hash:    hash,
		path:    path,
	}, nil
}

func Branch(gitRepoPath string) (string, error) {
	headPath := repository.HeadPath(gitRepoPath)
	content, err := os.ReadFile(headPath)
	if err != nil {
		return "", errors.New(
			`There is no branch in this repo,
            to create one use 'dzhigit update-ref'
            and then check it out using 'dzhigit checkout' `,
		)
	}
	branch := branchNameFromHead(string(content))
	return branch, nil
}

func Checkout(
	gitRepoPath string,
	branchName string,
	objPath string,
	reader repository.FileReader,
	formatter repository.GitFileFormatter,
) error {
	headsPath := repository.HeadsPath(gitRepoPath)
	pathToBranch := headsPath + branchName
	if !repository.Exists(pathToBranch) {
		return errors.New(
			fmt.Sprintf(
				"error branch with name '%s' doesn't exist",
				branchName,
			),
		)
	}
	treeHash, err := treeHashFromBranch(pathToBranch, objPath, formatter, reader)
	if err != nil {
		return err
	}
	err = checkoutRecursively(treeHash, "", reader, objPath, formatter)
	if err != nil {
		return err
	}
	head := repository.HeadPath(gitRepoPath)
	f, err := os.OpenFile(head, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(headContent(branchName)))
	return err
}

func checkoutRecursively(
	treeHash repository.Hash,
	rootPath string,
	reader repository.FileReader,
	objPath string,
	formatter repository.GitFileFormatter,
) error {
	//queue of inner trees
	queue := []checkoutTuple{}
	rawContent, err := reader(treeHash.Path(objPath))
	if err != nil {
		return err
	}
	deser, err := formatter.Deserialize(rawContent)
	lines := strings.Split(string(deser.Content), "\n")
	for _, line := range lines {
		treeEntry, err := newTreeEntry(line)
		if err != nil {
			return err
		}
		//if tree then store in queue and proceed later
		if treeEntry.objType == repository.TREE {
			queue = append(
				queue,
				checkoutTuple{
					treeHash: treeEntry.hash,
					path:     treeEntry.path,
				},
			)
		} else {
			//else override content right away
			data, err := reader(treeEntry.hash.Path(objPath))
			if err != nil {
				return err
			}
			err = os.WriteFile(rootPath+treeEntry.path, data, 0755)
			if err != nil {
				return err
			}
		}
	}
	for _, tuple := range queue {
		err := checkoutRecursively(
			tuple.treeHash,
			rootPath+tuple.path+"/",
			reader,
			objPath,
			formatter,
		)
		if err != nil {
			return err
		}
	}
	return nil
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

//creates a new branch
func UpdateRef(
	hash repository.Hash, //commit hash
	writer io.Writer, //writer to save a hash into commit file
	objPath string,
	reader repository.FileReader, //reader to read a hash
	formatter repository.GitFileFormatter,
) error {
	objType, err := repository.TypeByHash(objPath, hash, reader, formatter)
	if err != nil {
		return err
	}
	if objType != repository.COMMIT {
		return errors.New(
			fmt.Sprintf(
				"Wrong object type 'commit' expected, got %s",
				objType,
			),
		)
	}
	_, err = writer.Write([]byte(hash))
	return err
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
						commit.treeHash,
					),
				)
		}
	}
	return createCommitObject(
		commit,
		fileFormatter,
	)
}

//create a tree object from entries saved in index
func WriteTree(
	indexLines []string, //list of entries stored in index file
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

//cat object by given hash
func GitCat(
	hash repository.Hash,
	fileFormatter repository.GitFileFormatter,
	objPath string,
	reader repository.FileReader,
) (*repository.DeserializedGitObject, error) {
	if !repository.Exists(hash.Path(objPath)) {
		return nil, errors.New(fmt.Sprintf("File with hash %s doesn't exist", hash))
	}
	data, err := reader(hash.Path(objPath))
	if err != nil {
		return nil, err
	}
	return fileFormatter.Deserialize(data)
}

//Adds a new entry into an index file
func UpdateIndex(index repository.IndexEntry, indexPath string) error {
	f, err := os.OpenFile(
		indexPath,
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}
	defer f.Close()
	builder := strings.Builder{}
	s := bufio.NewScanner(f)
	indexEmpty := true
	foundDuplicate := false
	for s.Scan() {
		indexEmpty = false
		text := s.Text()
		parsedIndex, err := repository.ParseLineToIndex(text)
		if err != nil {
			return err
		}
		if index.Path() == parsedIndex.Path() {
			foundDuplicate = true
			builder.WriteString(index.String() + "\n")
		} else {
			builder.WriteString(parsedIndex.String() + "\n")
		}
	}
	if indexEmpty || !foundDuplicate {
		builder.WriteString(index.String() + "\n")
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(builder.String()))
	return err
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

//string that shows how a reference to a tree is stored inside of a tree file
func treeLine(hash repository.Hash, dir string) string {
	return fmt.Sprintf("040000 tree %s\t%s\n", hash, dir)
}
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
func branchNameFromHead(head string) string {
	parts := strings.Split(head, "/")
	return parts[len(parts)-1]
}

//TODO: introduce branch type ?
func treeHashFromBranch(
	pathToBranch string,
	objPath string,
	fileFormatter repository.GitFileFormatter,
	reader repository.FileReader,
) (repository.Hash, error) {
	rawContent, err := reader(pathToBranch)
	if err != nil {
		return "", err
	}
	deser, err := fileFormatter.Deserialize(rawContent)
	if err != nil {
		return "", err
	}
	commitHash, err := repository.NewHash(deser.Content)
	if err != nil {
		return "", err
	}
	rawContent, err = reader(commitHash.Path(objPath))
	if err != nil {
		return "", err
	}
	deser, err = fileFormatter.Deserialize(rawContent)
	if err != nil {
		return "", err
	}
	treeParts := strings.Split(string(deser.Content), "\n")[0]
	treeHash, err := repository.NewHash(strings.Split(treeParts, " ")[1])
	if err != nil {
		return "", err
	}
	return treeHash, nil
}

//content that will be stored in HEAD file
//Example :"refs: refs/heads/master"
func headContent(branchName string) string {
	return fmt.Sprintf("refs: %s", repository.PathToBranch(branchName))
}
