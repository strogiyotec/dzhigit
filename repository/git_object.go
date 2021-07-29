package repository

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
)

type GitObjectType string

const (
	BLOB GitObjectType = "blob"
	TREE               = "tree"
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
	Content []byte
}

type SerializedGitObject struct {
	Hash    string
	content []byte
}

func (obj *DefaultGitFileFormatter) Deserialize(data []byte) (*DeserializedGitObject, error) {
	spaceIndex := strings.Index(string(data), " ")
	objType, err := AsGitObjectType(string(data[0:spaceIndex]))
	if err != nil {
		return nil, err
	}
	nullIndex := strings.Index(string(data), "\x00")
	length := binary.BigEndian.Uint32(data[spaceIndex:nullIndex])
	if length != uint32(len(data)-nullIndex-1) {
		return nil,
			errors.New(
				fmt.Sprintf(
					`Invalid git object for deserialization , 
                    the length in the head is %d , expected %d\n`,
					length,
					len(data)-nullIndex-1,
				),
			)
	}
	return &DeserializedGitObject{
		objType: objType,
		Content: data[nullIndex+1:],
	}, nil
}

func (obj *DefaultGitFileFormatter) Serialize(data []byte, objType GitObjectType) (*SerializedGitObject, error) {
	header := header(data, objType)
	result := append(header, data...)
	hash, err := hashed(result)
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
		return errors.New("Hash already exists")
	}
}

//Git uses first two hash characters as a directory
//in order to decrease amount of files per directory
//some OS have limitations on amount of files per dir
func BlobDirWithFileName(hash string) (string, string) {
	return hash[0:2], hash[2:]
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

func hashed(data []byte) ([]byte, error) {
	hash := sha1.New()
	_, err := hash.Write(data)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil

}

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
