package repository

import (
	"crypto/sha1"
	"fmt"
	"testing"
)

func Test_SerializeAndDeserialize(t *testing.T) {
	content := "Hello world"
	fileFormatter := DefaultGitFileFormatter{}
	serialized, err := fileFormatter.Serialize([]byte(content), BLOB)
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if len(serialized.Hash) != sha1.Size*2 {
		t.Fatalf("Invalid hash length, expected %d , got %d ", sha1.Size*2, len(serialized.Hash))
	}
	deserialized, err := fileFormatter.Deserialize(serialized.content)
	if err != nil {
		t.Fatal(err)
	}
	if deserialized.objType != BLOB {
		t.Fatalf("Wrong object type , blog expected, got %s", deserialized.objType)
	}
	if content != deserialized.Content {
		t.Fatalf("Wrong content , expected '%s' , got '%s'", content, deserialized.Content)
	}

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
	_, err := AsGitObjectType("bla-bla")
	if err == nil {
		t.Fatal("Should not parse invalid git object type")
	}
}

func TestNewHash(t *testing.T) {
	//sha1 of "Some random data"
	predefinedHash := "3b0af1dd47d543b2166440b83bbf0ed0235173d8"
	hashBytes, err := generateHash([]byte("Some random data"))
	if err != nil {
		t.Error(err)
	}
	hashStr := fmt.Sprintf("%x", hashBytes)
	if hashStr != predefinedHash {
		t.Errorf("Wrong sha1 hash, given %s, expected %s", hashStr, predefinedHash)
	}
	hash, err := NewHash(hashStr)
	if err != nil {
		t.Error(err)
	}
	if hash.Dir() != predefinedHash[0:2] {
		t.Errorf("Wrong dir from hash, given %s, expected %s", hash.Dir(), predefinedHash[0:2])
	}
	if hash.FileName() != "0af1dd47d543b2166440b83bbf0ed0235173d8" {
		t.Errorf(
			"Wrong filename from hash, given %s, expected %s",
			hash.FileName(),
			predefinedHash[2:],
		)
	}
}
