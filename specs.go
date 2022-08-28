package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
		elements := make([]string, len(self.Elements))
		for i := 0; i < len(self.Elements); i++ {
			elements[i] = strings.ReplaceAll(self.Elements[i], " ", "\\ ")
		}
		builder.WriteString(fmt.Sprintf(":(%s)", strings.Join(elements, " ")))
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

const DESC_NEW_LINES = "\n                                   "

func printHelp(cmds *map[string]CmdSpec) {
	fmt.Println("Commands:\n\t")
	for name, spec := range *cmds {
		fmt.Printf("\t- %-25s%s\n", name, strings.Replace(spec.Description, "\n", DESC_NEW_LINES, -1))
	}
}
