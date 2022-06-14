package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

func logError(str string, params ...any) {
	fmt.Fprintf(os.Stderr, str, params...)
}

func loadSpec(conf *map[string]CmdSpec, name string) {
	found, err := Exists(name)
	if found && err == nil {
		yamlFile, err := ioutil.ReadFile(name)
		if err != nil {
			logError(fmt.Sprintf("Error reading: %s", name), err)
		}
		err = yaml.Unmarshal(yamlFile, conf)
		if err != nil {
			logError(fmt.Sprintf("Error parsing: %s", name), err)
		}
	}
}
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
	/*
		conf["zsh-autocomplete"] = CmdSpec{
			Description: "Generates the 'cb' autocomplete file for zsh",
		}
	*/
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

const DESC_NEW_LINES = "\n                                   "

func printHelp(cmds *map[string]CmdSpec) {
	fmt.Println("Commands:\n\t")
	for name, spec := range *cmds {
		fmt.Printf("\t- %-25s%s\n", name, strings.Replace(spec.Description, "\n", DESC_NEW_LINES, -1))
	}
}

const (
	CBCommand     = 1
	ScriptCommand = 3
	NIL           = "_THIS_VALUE_WAS_NOT_SET"
)

type OptType string

const (
	Text   = "text"
	Enum   = "enum"
	File   = "file"
	Folder = "folder"
)

type OptSpec struct {
	Description string
	Type        OptType
	Elements    []string
	Path        string
	Pattern     string
	Default     *string
}

func (self *OptSpec) description(extended bool) string {
	var builder strings.Builder
	builder.WriteString(self.Description)
	if self.Default == nil {
		builder.WriteString(" (required)")
	} else {
		builder.WriteString(fmt.Sprintf(" (%s)", *self.Default))
	}
	if extended {
		if len(self.Elements) > 0 {
			builder.WriteString(fmt.Sprintf(" [%s]", strings.Join(self.Elements, ", ")))
		}
		if len(self.Pattern) > 0 {
			builder.WriteString(fmt.Sprintf("[pattern:%s]", self.Pattern))
		}
		if len(self.Path) > 0 {
			builder.WriteString(fmt.Sprintf("[path:%s]", self.Path))
		}
	}
	return builder.String()
}

func (self OptSpec) toArgumentSpec(name string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("--%s", name))
	builder.WriteString(fmt.Sprintf("[%s]", self.description(false)))
	builder.WriteString(fmt.Sprintf(":%s", name))
	switch self.Type {
	case Text:
		// do nothing
	case Enum:
		builder.WriteString(fmt.Sprintf(":(%s)", strings.Join(self.Elements, " ")))
	case File:
		builder.WriteString(":_files")
		if len(self.Path) > 0 {
			builder.WriteString(" -W ")
			builder.WriteString(self.Path)
		}
		if len(self.Pattern) > 0 {
			builder.WriteString(" -g ")
			builder.WriteString(fmt.Sprintf("\"%s\"", self.Pattern))
		}
	case Folder:
		builder.WriteString(":_files -/")
		if len(self.Path) > 0 {
			builder.WriteString(" -W ")
			builder.WriteString(self.Path)
		}
		if len(self.Pattern) > 0 {
			builder.WriteString(" -g ")
			builder.WriteString(fmt.Sprintf("\"%s\"", self.Pattern))
		}
	default:
		panic(fmt.Sprintf("Unmanaged type [%s] for opt [%s]", self.Type, name))
	}
	return builder.String()
}

func (self *OptSpec) Help(name string) string {
	return fmt.Sprintf("--%-25s%s", name, self.description(true))
}

type FlagSpec struct {
	Description string
	Values      map[string]string
}

func (self *FlagSpec) ToArgumentSpec(name string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("-%s", name))
	if len(self.Description) > 0 {
		builder.WriteString(fmt.Sprintf("[%s]", self.Description))
	}
	return builder.String()
}

func (self *FlagSpec) Help(name string) string {
	return fmt.Sprintf("-%-25s%s | variables: %v", name, self.Description, self.Values)
}

type CmdSpec struct {
	Description string
	Opts        map[string]OptSpec
	Flags       map[string]FlagSpec
	Exec        string
}

func (self *CmdSpec) Help(name string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Command: %s\n", name))
	builder.WriteString(self.Description)
	builder.WriteString("\n\n")
	builder.WriteString(fmt.Sprintf("$ %s", self.Exec))
	if len(self.Opts) > 0 {
		builder.WriteString("\n\nArguments:")
		for name, opt := range self.Opts {
			builder.WriteString("\n")
			builder.WriteString(opt.Help(name))
		}
	}
	if len(self.Flags) > 0 {
		builder.WriteString("\n\nFlags:")
		for name, opt := range self.Flags {
			builder.WriteString("\n")
			builder.WriteString(opt.Help(name))
		}
	}
	return builder.String()
}
func (self *CmdSpec) GetArgumentsSpec() []string {
	var arguments []string
	for name, optSpec := range self.Opts {
		arguments = append(arguments, fmt.Sprintf("'%s'", optSpec.toArgumentSpec(name)))
	}
	for name, flagSpec := range self.Flags {
		arguments = append(arguments, fmt.Sprintf("'%s'", flagSpec.ToArgumentSpec(name)))
	}
	return arguments
}
func (self *CmdSpec) Run() {
	var arguments = make(map[string]*string)
	var flags = make(map[string]*bool)
	for name, opt := range self.Opts {
		var def string
		if opt.Default != nil {
			def = *opt.Default
		} else {
			def = NIL
		}
		arguments[name] = flag.String(name, def, "")
	}
	for name := range self.Flags {
		flags[name] = flag.Bool(name, false, "")
	}

	flag.Parse()

	cmd := exec.Command(os.Getenv("SHELL"), "-c", self.Exec)
	cmd.Env = os.Environ()
	// Set options
	for name, opt := range self.Opts {
		var value = *arguments[name]
		if opt.Default == nil && value == NIL {
			fmt.Fprintf(os.Stderr, "Missing value for: --%s ", name)
			os.Exit(-1)
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
		}
	}
	// Set flags
	for name, flag := range self.Flags {
		if *flags[name] {
			for n, v := range flag.Values {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", n, v))
			}
		}
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}
func Exists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
