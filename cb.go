package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	if os.Getenv("CB_DEBUG") == "true" {
		file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
		}
	} else {
		log.SetOutput(ioutil.Discard)
	}

	var conf = make(map[string]CmdSpec)
	loadSpec(&conf, os.Getenv("HOME")+"/.commands.yaml")
	loadSpec(&conf, "./.commands.yaml")
	log.Println("args:", os.Args)

	conf["help"] = createHelp(&conf)
	conf["zsh-autocomplete"] = zshAutocomplete()
	if len(os.Args) < 2 {
		printHelp(&conf)
	} else if os.Args[1] == "completion" {
		log.Println("Cmd list")
		keys := make([]string, 0, len(conf))
		for k := range conf {
			keys = append(keys, k)
		}
		fmt.Printf("'1:cmd:(%s)' ", strings.Join(keys, " "))

		if len(os.Args) > 3 {
			var command CmdSpec = conf[os.Args[ScriptCommand]]
			for _, s := range command.GetArgumentsSpec() {
				fmt.Print(s + " ")
			}
		}
	} else {
		if len(os.Args) < 2 {
			printHelp(&conf)
		} else {
			var name = os.Args[1]
			os.Args = os.Args[1:] // Shift args for flag.Parse to work

			if name == "help" {
				if len(os.Args) == 2 {
					printHelp(&conf)
				} else {
					var commandForHelp string
					flag.StringVar(&commandForHelp, "command", "help", "The name of the command")
					flag.Parse()
					cmd, found := conf[commandForHelp]
					if found {
						fmt.Println(cmd.Help(commandForHelp))
					} else {
						printHelp(&conf)
					}
				}
			} else {
				command, found := conf[name]
				if found {
					command.Run()
				} else {
					printHelp(&conf)
				}
			}
		}
	}
}

const (
	CBCommand     = 1
	ScriptCommand = 3
	NIL           = "_THIS_VALUE_WAS_NOT_SET"
)
