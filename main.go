package main

import (
	"fmt"

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
	default:
		fmt.Println("Default")
	}
}
