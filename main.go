package main

import (
	"fmt"
	"io/ioutil"

	"github.com/alecthomas/kong"
	"github.com/strogiyotec/dzhigit/cli"
	"github.com/strogiyotec/dzhigit/repository"
)

func main() {
	ctx := kong.Parse(&cli.Git)
	switch ctx.Command() {
	case "add <files>":
		path := repository.DefaultPath()
		command, err := cli.NewGitAdd(cli.Git.Add.Files, path)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			command.Add()
		}
	case "init":
		{
			path := repository.DefaultPath()
			err := repository.Init(path)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Initialize new dzhigit repository")
			}
		}
	case "hash-object <file>":
		{
			if true {
				fmt.Println(cli.Git.HashObject.Type)
				return
			}
			gitFile := repository.DefaultGitFileFormatter{}
			content, err := ioutil.ReadFile(cli.Git.HashObject.File)
			if err != nil {
				fmt.Printf(
					"Error during reading file for hashing %s ",
					err.Error(),
				)
				return
			}
			objType, _ := repository.AsGitObjectType(cli.Git.HashObject.Type)
			serialized, err := gitFile.Serialize(content, objType)
			if err != nil {
				fmt.Printf(
					"Error during serializing file for hashing %s ",
					err.Error(),
				)
				return
			}
			if cli.Git.HashObject.Write {
				path := repository.DefaultPath()
				if !repository.Exists(path) {
					fmt.Println("Dzhigit repository doesn't exist")
					return
				}
				objPath := repository.ObjPath(path)
				err = gitFile.Save(serialized, objPath)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}
			fmt.Println(serialized.Hash)
		}
	case "cat-file <hash>":
		{
			path := repository.DefaultPath()
			if !repository.Exists(path) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			gitFile := &repository.DefaultGitFileFormatter{}
			objPath := repository.ObjPath(path)
			var reader cli.FileReader = func(path string) ([]byte, error) {
				return ioutil.ReadFile(path)
			}
			deser, err := cli.GitCat(cli.Git.CatFile.Hash, gitFile, objPath, reader)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(deser.Content)
		}
	case "update-index <hash> <file> <mode>":
		{
			path := repository.DefaultPath()
			if !repository.Exists(path) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			objPath := repository.ObjPath(path)
			indexParams := cli.Git.UpdateIndex
			index, err := repository.NewIndex(indexParams.File, indexParams.Mode, indexParams.Hash, objPath)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(index.String())
			}
		}
	default:
		fmt.Println("Default")
	}
}
