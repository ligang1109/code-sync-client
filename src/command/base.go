package command

import (
	"code-sync-client/resource"
	"errors"
	"flag"
	"github.com/goinbox/golog"
	"strings"
)

type ICommand interface {
	Run(args []string) error
}

type newCommandFunc func() ICommand

var commandTable = make(map[string]newCommandFunc)

func register(name string, ncf newCommandFunc) {
	commandTable[name] = ncf
}

func NewCommandByName(name string) ICommand {
	ncf, ok := commandTable[name]
	if !ok {
		return nil
	}

	return ncf()
}

type runFunc func() error

type baseCommand struct {
	Fs *flag.FlagSet

	ExtArgs map[string]string

	mustHaveArgs map[string]bool
	existArgs    map[string]bool

	rf runFunc
}

func NewBaseCommand() *baseCommand {
	return &baseCommand{
		Fs:      new(flag.FlagSet),
		ExtArgs: make(map[string]string),

		mustHaveArgs: make(map[string]bool),
		existArgs:    make(map[string]bool),
	}
}

func (bc *baseCommand) AddMustHaveArgs(names ...string) *baseCommand {
	for _, name := range names {
		bc.mustHaveArgs[name] = true
	}

	return bc
}

func (bc *baseCommand) SetRunFunc(rf runFunc) *baseCommand {
	bc.rf = rf

	return bc
}

func (bc *baseCommand) Run(args []string) error {
	err := bc.parseArgs(args)
	if err != nil {
		return err
	}

	for name, _ := range bc.mustHaveArgs {
		_, ok := bc.existArgs[name]
		if !ok {
			return errors.New("Must have arg " + name)
		}
	}

	return bc.rf()
}

func (bc *baseCommand) parseArgs(args []string) error {
	err := bc.Fs.Parse(args)
	if err != nil {
		return err
	}

	bc.Fs.Visit(func(f *flag.Flag) {
		bc.existArgs[f.Name] = true
	})

	for _, str := range bc.Fs.Args() {
		item := strings.Split(str, "=")
		if len(item) == 2 {
			bc.ExtArgs[item[0]] = item[1]
			bc.existArgs[item[0]] = true
		}
	}

	return nil
}

func (bc *baseCommand) changeToConsoleLogger() {
	resource.AccessLogger = resource.NewLogger(golog.NewConsoleWriter())
}
