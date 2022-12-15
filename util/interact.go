package util

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"os"
	"path/filepath"
	"strconv"
)

type Interact struct {
}

var (
	hostQs = []*survey.Question{
		{
			Name: "host",
			Prompt: &survey.Input{
				Message: "Please provide the database address, the default is \"localhost\" ",
				Default: "localhost",
			},
			Validate: func(ans interface{}) error {
				if err := survey.Required(ans); nil != err {
					return err
				}
				if _, ok := ans.(string); !ok {
					return errors.New("Input type error")
				}
				return nil
			},
		},
	}
	portQs = []*survey.Question{
		{
			Name: "port",
			Prompt: &survey.Input{
				Message: "Please provide the port of database, the default is \"3306\" ",
				Default: "3306",
			},
			Validate: func(ans interface{}) error {
				if err := survey.Required(ans); nil != err {
					return err
				}
				if v, ok := ans.(string); !ok {
					return errors.New("Input type error")
				} else {
					if _v, err := strconv.ParseUint(v, 10, 16); nil != err {
						return errors.New("Input type error")
					} else {
						if 0 >= _v || 65535 < _v {
							return errors.New("Port out of bounds")
						}
					}

				}
				return nil
			},
		},
	}
	userQs = []*survey.Question{
		{
			Name: "user",
			Prompt: &survey.Input{
				Message: "Please provide the user of database, the default is \"root\" ",
				Default: "root",
			},
			Validate: func(ans interface{}) error {
				if err := survey.Required(ans); nil != err {
					return err
				}
				if _, ok := ans.(string); !ok {
					return errors.New("Input type error")
				}
				return nil
			},
		},
	}
	passwdQs = []*survey.Question{
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Please provide the password of database, the default is \"\" ",
			},
		},
	}
	databaseQs = []*survey.Question{
		{
			Name: "database",
			Prompt: &survey.Input{
				Message: "Please provide the database name ",
				Default: "",
			},
			Validate: survey.Required,
		},
	}
	packageQs = []*survey.Question{
		{
			Name: "package",
			Prompt: &survey.Input{
				Message: "Please provide the java package ",
				Default: "",
			},
			Validate: survey.Required,
		},
	}
	exportPathQs = []*survey.Question{
		{
			Name: "exportPath",
			Prompt: &survey.Input{
				Message: "Please provide the export directory path ",
				Default: "",
			},
			Validate: func(ans interface{}) error {
				if v, ok := ans.(string); !ok {
					return errors.New("Input type error, ExportPath")
				} else {
					if _, err := filepath.Abs(v); nil != err {
						return err
					}
				}
				return nil
			},
		},
	}
)

func (interact *Interact) AskDBHost() string {
	answers := struct {
		Host string `survey:"host"`
	}{}
	err := survey.Ask(hostQs, &answers)
	if nil != err {
		return "localhost"
	}
	return answers.Host
}

func (Interact *Interact) AskDBPort() uint16 {
	answers := struct {
		Port uint16 `survey:"port"`
	}{}
	if err := survey.Ask(portQs, &answers); nil != err {
		return 3306
	}
	return answers.Port
}

func (interact *Interact) AskDBUser() string {
	answers := struct {
		User string `survey:"user"`
	}{}
	err := survey.Ask(userQs, &answers)
	if nil != err {
		return "root"
	}
	return answers.User
}

func (interact *Interact) AskDBPassword() string {
	answers := struct {
		Password string `survey:"password"`
	}{}
	err := survey.Ask(passwdQs, &answers)
	if nil != err {
		return ""
	}
	return answers.Password
}

func (interact *Interact) AskDBName() string {
	answers := struct {
		DbName string `survey:"database"`
	}{}
	err := survey.Ask(databaseQs, &answers)
	if nil != err {
		return ""
	}
	return answers.DbName
}

func (interact *Interact) AskPackage() string {
	answers := struct {
		Value string `survey:"package"`
	}{}
	err := survey.Ask(packageQs, &answers)
	if nil != err {
		return ""
	}
	return answers.Value
}

func (interact *Interact) AskExportPath() string {
	answers := struct {
		Value string `survey:"exportPath"`
	}{}
	err := survey.Ask(exportPathQs, &answers)
	if nil != err || "" == answers.Value {
		wd, _ := os.Getwd()
		return wd
	}
	return answers.Value
}
