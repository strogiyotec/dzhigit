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
	CatFile struct {
		Hash string `arg name:"hash" help:"object hash"`
	} `cmd help:"Print the content of a blob by hash"`
	UpdateIndex struct {
		Hash string `arg name:"hash" help:"hash" `
		File string `arg name:"file" help:"path to file to save in index" `
		Mode string `arg name:"mode" help:"file mode" enum:"100644,100755" default"100644"`
	} `cmd help:"Update index"`
	LsTree struct {
	} `cmd help:"Print index content"`
	WriteTree struct {
	} `cmd help:"Create a tree object from index file"`
	CommitTree struct {
		Message string `help:"Commit message" short:"m" required:""`
		Parent  string ` help:"hash of a parent commit" short:"p" default:"" `
		Hash    string `arg name:"hash" help:"hash of a tree object" `
	} `cmd help:"Create a commit object"`
}
