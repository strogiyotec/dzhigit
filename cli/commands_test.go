package cli

import (
	"testing"
)

func TestNewUser(t *testing.T) {
	content := `{"name":"Almas","email":"almas337519@gmail.com"}`
	user, err := NewUser([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if user.Email != "almas337519@gmail.com" {
		t.Fatal("Wrong email for deserialized user")
	}
	if user.Name != "Almas" {
		t.Fatal("Wrong name for deserialized user")
	}
}
