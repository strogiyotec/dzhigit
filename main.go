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
	case "init":
		{
			path := repository.DefaultPath()
			json, err := cli.DefaultGitUserAsJson()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			err = repository.Init(path, json)
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
					"Error during reading file for hashing\n %s ",
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
			hash, err := repository.NewHash(cli.Git.CatFile.Hash)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			deser, err := cli.GitCat(hash, gitFile, objPath, repository.Reader)
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
			//TODO: if file with hash already exists,override hash
			mode, err := repository.AsMode(indexParams.Mode)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			hash, err := repository.NewHash(indexParams.Hash)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			index, err := repository.NewIndex(indexParams.File, mode, hash, objPath)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				indexPath := repository.IndexPath(gitRepoPath)
				err := cli.UpdateIndex(*index, indexPath)
				if err != nil {
					fmt.Println(err.Error())
				} else {
					fmt.Printf("File %s was saved in index\n", indexParams.File)
				}
			}
		}
	case "ls-tree":
		{
			//TODO: move this check above all switch cases
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
			indexContent, err := ioutil.ReadFile(indexPath)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			lines := strings.Split(strings.TrimSpace(string(indexContent)), "\n")
			objPath := repository.ObjPath(gitRepoPath)
			tree, err := cli.WriteTree(
				lines,
				objPath,
				&repository.DefaultGitFileFormatter{},
			)
			if err != nil {
				fmt.Println(err.Error())
				return
			} else {
				fmt.Println(tree.Hash)
			}
		}
	case "commit-tree <hash>":
		{
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			content, err := os.ReadFile(repository.ConfigPath(gitRepoPath))
			if err != nil {
				fmt.Printf("Error reading a config file %s", err.Error())
				return
			}
			user, err := cli.NewUser(content)
			if err != nil {
				fmt.Printf(
					"Error reading a user's data from config file %s",
					err.Error(),
				)
				return
			}
			time := cli.CurrentTime()
			treeHash, err := repository.NewHash(cli.Git.CommitTree.Hash)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			var parentHash repository.Hash
			if len(cli.Git.CommitTree.Parent) != 0 {
				parentHash, err = repository.NewHash(cli.Git.CommitTree.Parent)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}
			commit := cli.NewCommit(
				treeHash,
				cli.Git.CommitTree.Message,
				parentHash,
			)
			repo := &repository.DefaultGitFileFormatter{}
			objPath := repository.ObjPath(gitRepoPath)
			ser, err := cli.CommitTree(
				*commit,
				*time,
				objPath,
				*user,
				repo,
				repository.Reader,
			)
			if err != nil {
				fmt.Println(err.Error())
				return
			} else {
				err := repo.Save(ser, objPath)
				if err != nil {
					fmt.Printf(
						"Error saving a commit object %s",
						err.Error(),
					)
					return
				} else {
					fmt.Println(ser.Hash)
				}
			}
		}
	case "update-ref <name> <hash>":
		{
			options := cli.Git.UpdateRef
			formatter := repository.DefaultGitFileFormatter{}
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			objPath := repository.ObjPath(gitRepoPath)
			treeHash, err := repository.NewHash(options.Hash)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			headsPath := repository.HeadsPath(gitRepoPath)
			f, err := os.OpenFile(
				headsPath+options.Name,
				os.O_RDWR|os.O_CREATE|os.O_TRUNC,
				0755,
			)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			err = cli.UpdateRef(treeHash, f, objPath, repository.Reader, &formatter)
			if err != nil {
				fmt.Println(err.Error())
				return
			} else {
				fmt.Printf("Branch %s was created", options.Name)
			}
		}
	case "checkout <branch>":
		{
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			options := cli.Git.Checkout
			err := cli.Checkout(
				gitRepoPath,
				options.Branch,
			)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("Branch %s was checkout\n", options.Branch)
		}
	case "branch":
		{
			gitRepoPath := repository.DefaultPath()
			if !repository.Exists(gitRepoPath) {
				fmt.Println("Dzhigit repository doesn't exist")
				return
			}
			branch, err := cli.Branch(gitRepoPath)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Printf("* %s\n", branch)
		}
	default:
		fmt.Println("Default")
	}
}
