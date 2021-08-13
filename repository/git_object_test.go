package repository

import (
	"crypto/sha1"
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
