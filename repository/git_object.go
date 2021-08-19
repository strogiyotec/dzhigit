package repository

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type GitObjectType string

const (
	BLOB   GitObjectType = "blob"
	TREE                 = "tree"
	COMMIT               = "commit"
)

//reader to read a file content by path
//use first class function to improve testability
type FileReader func(path string) ([]byte, error)

//default reader
var Reader FileReader = func(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

type GitFileFormatter interface {
	Serialize(data []byte, objType GitObjectType) (*SerializedGitObject, error)
	Deserialize(data []byte) (*DeserializedGitObject, error)
	Save(serialized *SerializedGitObject, path string) error
}

type TreeAlreadyExistError struct {
	message string
}

type DefaultGitFileFormatter struct {
}

type DeserializedGitObject struct {
	ObjType GitObjectType
	Content string
}

type SerializedGitObject struct {
	Hash    Hash
	Content []byte
}

type Hash string

func NewHash(hash string) (Hash, error) {
	if len(hash) != sha1.Size*2 {
		return "",
			errors.New(
				fmt.Sprintf(
					"Wrong hash length ,should be %d, got %d",
					sha1.Size*2,
					len(hash),
				),
			)
	}
	return Hash(hash), nil
}

func (h Hash) Dir() string {
	return string(h)[0:2]
}

//returns the whole directory to the file with given hash
func (h Hash) Path(objPath string) string {
	return objPath + h.Dir() + "/" + h.FileName()
}

func (h Hash) FileName() string {
	return string(h)[2:]
}

func (err *TreeAlreadyExistError) Error() string {
	return err.message
}

func NewTreeAlreadyExistError(message string) *TreeAlreadyExistError {
	return &TreeAlreadyExistError{
		message: message,
	}
}

func (obj *DefaultGitFileFormatter) Deserialize(data []byte) (*DeserializedGitObject, error) {
	unzipped, err := unzipped(data)
	unzippedContent := string(unzipped)
	if err != nil {
		return nil, err
	}
	return newDeserializedObj(unzippedContent)
}

func (obj *DefaultGitFileFormatter) Serialize(data []byte, objType GitObjectType) (*SerializedGitObject, error) {
	header := header(data, objType)
	result := append(header, data...)
	hash, err := GenerateHash(result)
	if err != nil {
		return nil, err
	}
	zipped, err := zipped(result)
	if err != nil {
		return nil, err
	}
	return &SerializedGitObject{
		Hash:    Hash(hash),
		Content: zipped,
	}, nil
}

func (obj *DefaultGitFileFormatter) Save(serialized *SerializedGitObject, objPath string) error {
	fullPath := objPath + serialized.Hash.Dir()
	//dir may already exist
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		os.Mkdir(fullPath, 0755)
	}
	file, err := os.Create(fullPath + "/" + serialized.Hash.FileName())
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.Write(serialized.Content)
	if err != nil {
		return err
	}
	return writer.Flush()
}

//Get type of a object by given hash
func TypeByHash(
	objPath string,
	hash Hash,
	reader FileReader,
	fileFormatter GitFileFormatter,
) (GitObjectType, error) {
	if !Exists(objPath + hash.Dir() + "/" + hash.FileName()) {
		return "", errors.New(fmt.Sprintf("Object with hash %s doesn't exist", hash))
	}
	data, err := reader(objPath + hash.Dir() + "/" + hash.FileName())
	if err != nil {
		return "", err
	}
	deser, err := fileFormatter.Deserialize(data)
	if err != nil {
		return "", err
	}
	return deser.ObjType, nil
}

func newDeserializedObj(content string) (*DeserializedGitObject, error) {
	spaceIndex := strings.Index(content, " ")
	objType, err := AsGitObjectType(content[0:spaceIndex])
	if err != nil {
		return nil, err
	}
	nullIndex := strings.Index(content, "\x00")
	length, err := strconv.Atoi(content[spaceIndex+1 : nullIndex])
	if err != nil {
		return nil, err
	}
	if length != len(content)-nullIndex-1 {
		return nil,
			errors.New(
				fmt.Sprintf(
					`Invalid git object for deserialization , 
                    the length in the head is %d , expected %d\n`,
					length,
					len(content)-nullIndex-1,
				),
			)
	}
	return &DeserializedGitObject{
		ObjType: objType,
		Content: content[nullIndex+1:],
	}, nil
}

func unzipped(data []byte) ([]byte, error) {
	buffer := bytes.NewReader(data)
	reader, err := zlib.NewReader(buffer)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var out bytes.Buffer
	_, err = io.Copy(&out, reader)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func zipped(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	w := zlib.NewWriter(&buffer)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	w.Close()
	return buffer.Bytes(), nil
}

func GenerateHash(data []byte) (string, error) {
	hash := sha1.New()
	_, err := hash.Write(data)
	if err != nil {
		return "", err
	}
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))
	return hashStr, nil

}

//[type] [length][null]
func header(data []byte, objType GitObjectType) []byte {
	return []byte(
		fmt.Sprintf(
			"%s %d\x00",
			objType,
			len(data),
		),
	)
}

//convert given string to git object type
func AsGitObjectType(str string) (GitObjectType, error) {
	switch str {
	case "blob":
		return BLOB, nil
	case "tree":
		return TREE, nil
	case "commit":
		return COMMIT, nil
	default:
		return "",
			errors.New(
				fmt.Sprintf(
					"%s is not a valid git object type",
					str,
				),
			)
	}
}
