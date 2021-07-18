package repository

import (
	"errors"
	"os"
)

const (
	Objects     = "/objects/"
	Refs        = "/refs/"
	Head        = "Head"
	Config      = "Config"
	Description = "Description"
)

func DefaultPath() string {
	path, _ := os.Getwd()
	return path + "/.dzhigit/"
}

//Check if repository exists
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

//create git repository
func Init(path string) error {
	if !Exists(path) {
		return initRepo(path)
	} else {
		return errors.New("Dzhigit repository already exists")
	}
}

func initRepo(path string) error {
	err := os.Mkdir(path, 0755)
	if err != nil {
		return err
	}
	//Create Objects dir
	err = os.Mkdir(path+Objects, 0755)
	if err != nil {
		return err
	}
	//Create Refs dir
	err = os.Mkdir(path+Refs, 0755)
	if err != nil {
		return nil
	}
	//Create Config
	_, err = os.Create(path + Config)
	if err != nil {
		return nil
	}
	//Create Description
	_, err = os.Create(path + Description)
	if err != nil {
		return nil
	}
	//Create Head
	_, err = os.Create(path + Head)
	if err != nil {
		return nil
	}
	return nil
}
