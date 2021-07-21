package repository

import (
	"crypto/sha1"
	"fmt"
	"testing"
)

func Test_SerializeAndDeserialize(t *testing.T) {
	fileFormatter := DefaultGitFileFormatter{}
	serialized, err := fileFormatter.Serialize([]byte("Hello world"), BLOB)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if len(serialized.hash) != sha1.Size {
		fmt.Println("Invalid hash length, has to be 40 ")
	}
	//TODO: test content of the serialized object
}

func Test_header(t *testing.T) {
	data := []byte("Hello world")
	header := header(data, BLOB)
	if string(header) != "blob 11\x00" {
		t.Fatal(
			"Invalid header",
		)
	}
}

func Test_asGitObjectType(t *testing.T) {
	_, err := asGitObjectType("bla-bla")
	if err == nil {
		t.Fatal("Should not parse invalid git object type")
	}
}
