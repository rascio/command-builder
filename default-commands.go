package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func createHelp(cmds *map[string]CmdSpec) CmdSpec {
	keys := make([]string, 0, len(*cmds))
	for k := range *cmds {
		keys = append(keys, k)
	}
	return CmdSpec{
		Description: "HELP!",
		Opts: map[string]OptSpec{
			"command": {
				Description: "Commands",
				Type:        Enum,
				Elements:    keys,
			},
		},
	}
}

func zshAutocomplete() CmdSpec {
	path, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return CmdSpec{
		Description: "Generate ZSH autocomplete",
		Exec: fmt.Sprintf(
			`echo '#compdef cb
eval "_arguments -s $(%s completion $words)"' > %s/_cb
			echo 'export fpath=($fpath %s) >> ~/.zshrc'`,
			path,
			filepath.Dir(path),
			filepath.Dir(path)),
	}
}
