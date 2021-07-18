package cli

var Git struct {
	Init struct {
	} `cmd help:"Init empty repository"`
	Add struct {
		Files []string `arg name:"files" help:"files to add" type:"path"`
	} `cmd help:"Add files."`
}
