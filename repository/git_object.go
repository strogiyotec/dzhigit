package repository

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
)

type GitFileFormatter interface {
	Serialize(data []byte, objType string) (*GitObject, error)
}

type DefaultGitFileFormatter struct {
}

type GitObject struct {
	hash    []byte
	content []byte
}

func (obj *DefaultGitFileFormatter) Serialize(data []byte, objType string) (*GitObject, error) {
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
	return &GitObject{
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

func header(data []byte, objType string) []byte {
	return []byte(
		fmt.Sprintf(
			"%s %d \x00",
			objType,
			len(data),
		),
	)
}
