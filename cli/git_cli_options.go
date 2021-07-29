package cli

var Git struct {
	Init struct {
	} `cmd help:"Init empty repository"`
	Add struct {
		Files []string `arg name:"files" help:"files to add" type:"path"`
	} `cmd help:"Add files."`
	HashObject struct {
		Write bool   `help:"Save the object" short:"w"`
		Type  string `help:"Type of object" enum:"blob,tree" default:"blob"`
		File  string `arg name:"file" help:"path to file to generate hash from" type :"path"`
	} `cmd help:"Get the hash of an object"`
}
