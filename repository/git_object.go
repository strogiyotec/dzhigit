package repository

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
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

type GitFileFormatter interface {
	Serialize(data []byte, objType GitObjectType) (*SerializedGitObject, error)
	Deserialize(data []byte) (*DeserializedGitObject, error)
	Save(serialized *SerializedGitObject, path string) error
}

type DefaultGitFileFormatter struct {
}

type DeserializedGitObject struct {
	objType GitObjectType
	Content string
}

type SerializedGitObject struct {
	Hash    string
	content []byte
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
	hash, err := generateHash(result)
	if err != nil {
		return nil, err
	}
	zipped, err := zipped(result)
	if err != nil {
		return nil, err
	}
	return &SerializedGitObject{
		Hash:    fmt.Sprintf("%x", hash),
		content: zipped,
	}, nil
}
func (obj *DefaultGitFileFormatter) Save(serialized *SerializedGitObject, path string) error {
	hashDir, fileName := BlobDirWithFileName(serialized.Hash)
	fullPath := path + hashDir
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		os.Mkdir(fullPath, 0755)
		file, err := os.Create(fullPath + "/" + fileName)
		if err != nil {
			return err
		}
		defer file.Close()
		writer := bufio.NewWriter(file)
		_, err = writer.Write(serialized.content)
		if err != nil {
			return err
		}
		return writer.Flush()
	} else {
		return errors.New(fmt.Sprintf("Hash %s already exists", serialized.Hash))
	}
}

//Git uses first two hash characters as a directory
//in order to decrease amount of files per directory
//some OS have limitations on amount of files per dir
func BlobDirWithFileName(hash string) (string, string) {
	return hash[0:2], hash[2:]
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
		objType: objType,
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

func generateHash(data []byte) ([]byte, error) {
	hash := sha1.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil

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
