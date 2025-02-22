package util

import (
    "errors"
    "fmt"
    "github.com/AlecAivazis/survey/v2"
    "github.com/AlecAivazis/survey/v2/terminal"
    "mybatis-export/config"
    "os"
    "path/filepath"
    "strconv"
    "strings"
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
                Message: "Please provide the java root package, e: com.google, com.github.aaa",
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
    tableNamesIsAllQs = &survey.Confirm{
        Message: "You did not provide a table name, are you processing all tables in the database?",
        Default: false,
    }
    tableNamesQs = []*survey.Question{
        {
            Name: "tableNames",
            Prompt: &survey.Input{
                Message: "Please provide some table names, How to have multiple values, please use spaces to separate",
            },
            Validate: survey.Required,
        },
    }
    entityPackageQs = &survey.Input{
        Message: "Please provide the entity package, do not need to include the root package. The default value is \"entity\"",
        Default: "entity",
    }
    mapperXmlPathQs = &survey.Input{
        Message: "Please provide mapper xml path, do not need to include the root path. The default value is \"resource\"",
        Default: "resource",
    }
    mapperPackageQs = &survey.Input{
        Message: "Please provide mapper package, do not need to include the root package. The default value is \"mapper\"",
        Default: "mapper",
    }
    queryPackageQs = &survey.Input{
        Message: "Please provide query package, do not need to include the root package. The default value is \"model.query\"",
        Default: "model.query",
    }
    tablePrefixsQs = []*survey.Question{
        {
            Name: "tablePrefixs",
            Prompt: &survey.Input{
                Message: "Please provide some table prefix list, How to have multiple values, please use \",\" to separate",
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
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return "localhost"
    }
    return answers.Host
}

func (Interact *Interact) AskDBPort() uint16 {
    answers := struct {
        Port uint16 `survey:"port"`
    }{}
    if err := survey.Ask(portQs, &answers); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
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
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
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
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
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
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
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
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return ""
    }
    return answers.Value
}

func (interact *Interact) AskExportPath() string {
    answers := struct {
        Value string `survey:"exportPath"`
    }{}
    err := survey.Ask(exportPathQs, &answers)
    if nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        wd, _ := os.Getwd()
        return wd
    }
    if "" == answers.Value {
        wd, _ := os.Getwd()
        return wd
    }
    return answers.Value
}

func (interact *Interact) AskIsAllTableOfDB() bool {
    ret := true
    if err := survey.AskOne(tableNamesIsAllQs, &ret); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return true
    }
    return ret
}

func (interact *Interact) AskTables() []string {
    answers := struct {
        Value string `survey:"tableNames"`
    }{}
    if err := survey.Ask(tableNamesQs, &answers); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return []string{}
    }
    return strings.Split(answers.Value, " ")
}

func (interact *Interact) AskEntityPackage() string {
    var entityPackage string
    if err := survey.AskOne(entityPackageQs, &entityPackage); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return "entity"
    }
    return entityPackage
}

func (interact *Interact) AskMapperPackage() string {
    var mapperPackage string
    if err := survey.AskOne(mapperPackageQs, &mapperPackage); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return "mapper"
    }
    return mapperPackage
}

func (interact *Interact) AskMapperXmlPath() string {
    var mapperXmlPath string
    if err := survey.AskOne(mapperXmlPathQs, &mapperXmlPath); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return "resource"
    }
    return mapperXmlPath
}

func (interact *Interact) AskQueryPackage() string {
    var queryPackage string
    if err := survey.AskOne(queryPackageQs, &queryPackage); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return "model.query"
    }
    return queryPackage
}

func (interact *Interact) AskIsOverwrite(what string) string {
    msg := fmt.Sprintf("The file \"%s\" already exists, whether to overwrite", what)
    overwriteQs := &survey.Select{
        Message: msg,
        Options: []string{"overwrite", "no", "overwrite all", "no all"},
        Default: "no",
    }
    var ret string = "no"
    if err := survey.AskOne(overwriteQs, &ret); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return ret
    }
    return ret
}

func (interact *Interact) AskTablePrefixs() []string {
    answers := struct {
        Value string `survey:"tablePrefixs"`
    }{}
    if err := survey.Ask(tablePrefixsQs, &answers); nil != err {
        if terminal.InterruptErr == err {
            Exit()
            os.Exit(0)
        }
        return []string{}
    }
    return strings.Split(answers.Value, ",")
}

func Exit() {
    if nil != config.DbIns {
        config.DbIns.Close()
    }
}
