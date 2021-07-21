package repository

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

type GitObjectType string

const (
	BLOB GitObjectType = "blob"
	TREE               = "tree"
)

//convert given string to git object type
func asGitObjectType(str string) (GitObjectType, error) {
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

type GitFileFormatter interface {
	Serialize(data []byte, objType string) (*SerializedGitObject, error)
	Deserialize(data []byte) (*DeserializedGitObject, error)
}

type DefaultGitFileFormatter struct {
}

type DeserializedGitObject struct {
	objType GitObjectType
	content []byte
}

func NewDeserializedGitObject(objType string, cotent []byte) {
}

type SerializedGitObject struct {
	hash    []byte
	content []byte
}

func (obj *DefaultGitFileFormatter) Deserialize(data []byte) (*DeserializedGitObject, error) {
	spaceIndex := strings.Index(string(data), " ")
	objType, err := asGitObjectType(string(data[0:spaceIndex]))
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
		content: data[nullIndex+1:],
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
		hash:    hash,
		content: zipped,
	}, nil

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
