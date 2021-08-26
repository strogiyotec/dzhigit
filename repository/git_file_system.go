package repository

import (
	"errors"
	"os"
)

const (
	Objects     = "/objects/"
	Refs        = "/refs"
	Heads       = "/heads/"
	Head        = "/HEAD"
	Config      = "/config.json"
	Description = "Description"
	Index       = "/index"
)

func DefaultPath() string {
	path, _ := os.Getwd()
	return path + "/.dzhigit"
}

func PathToBranch(branchName string) string {
	return Refs + Heads + branchName
}

//TODO:rename all path params to root
func HeadPath(path string) string {
	return path + Head
}

func HeadsPath(path string) string {
	return path + Refs + Heads
}

func ConfigPath(path string) string {
	return path + Config
}

func IndexPath(path string) string {
	return path + Index
}

func ObjPath(path string) string {
	return path + Objects
}

//Check if path exists
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

//create git repository, returns an error if already exists
func Init(path string, userJson []byte) error {
	if !Exists(path) {
		return initRepo(path, userJson)
	} else {
		return errors.New("dzhigit repository already exists")
	}
}

func initRepo(path string, userJson []byte) error {
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
	err = os.Mkdir(path+Refs+Heads, 0755)
	if err != nil {
		return nil
	}
	//Create Config
	config, err := os.Create(path + Config)
	if err != nil {
		return nil
	}
	defer config.Close()
	_, err = config.Write(userJson)
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
