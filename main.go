package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/strogiyotec/dzhigit/cli"
	"github.com/strogiyotec/dzhigit/repository"
)

func main() {
	ctx := kong.Parse(&cli.Git)
	switch ctx.Command() {
	case "add <files>":
		fmt.Println(cli.Git.Add.Files)
	case "init":
		{
			path, _ := os.Getwd()
			err := repository.Init(path + "/.dzhigit/")
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
