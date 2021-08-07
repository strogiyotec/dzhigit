package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
				fmt.Println(serialized.Hash)
				fmt.Println("The file with given hash was saved")
			} else {
				fmt.Println(serialized.Hash)
			}
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
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			objPath := repository.ObjPath(gitRepoPath)
			indexParams := cli.Git.UpdateIndex
			//TODO: check if already exists
			index, err := repository.NewIndex(indexParams.File, indexParams.Mode, indexParams.Hash, objPath)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				indexPath := repository.IndexPath(gitRepoPath)
				err := cli.UpdateIndex(*index, indexPath)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Printf("File %s was saved in index", indexParams.File)
				}
			}
		}
	case "ls-tree":
		{
			//TODO: move above switch case
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			indexPath := repository.IndexPath(gitRepoPath)
			file, err := os.Open(indexPath)
			if err != nil {
				fmt.Printf("Error opening index file %s", err.Error())
				return
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				fmt.Println(scanner.Text())
			}
		}
	case "write-tree":
		{
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			indexPath := repository.IndexPath(gitRepoPath)
			content, err := ioutil.ReadFile(indexPath)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			_, err = cli.WriteTree(lines)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	default:
		fmt.Println("Default")
	}
}
