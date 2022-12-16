/*
Copyright © 2022 BottleHe

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"mybatis-export/config"
	"mybatis-export/util"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var (
	host         string
	user         string
	password     string
	port         *uint16
	databaseName string
	tableNames   []string
	tablePrefix  string

	rootPath         string
	rootPackagePath  string
	entityPackage    string
	mapperXmlPath    string
	mapperPackage    string
	queryPackage     string
	queryRootPackage string
	allTable         *bool
	overwriteAll     *bool

	conflictOverwriteAll bool = false
	interact             util.Interact
)

type table struct {
	TableName string
	Comment   string
}

type column struct {
	Field     string
	DataType  string
	Index     string
	Comment   string
	IsPk      int
	IsIndex   int
	Property  string
	PropertyN string
	JdbcType  string
	JavaType  string
}

type TemplateData struct {
	Pk               string
	PkHump           string
	PackagePath      string
	TableNote        string
	TableName        string
	TableNameHump    string
	EntityPackage    string
	QueryPackage     string
	QueryRootPackage string
	MapperPackage    string
	Fields           []column
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mybatis-export [flags]",
	Short: "export mybatis project",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	PreRunE: func(cmd *cobra.Command, args []string) error {
		//fmt.Printf("Args:  \n"+
		//	"host: %v\n"+
		//	"port: %v\n"+
		//	"user: %v\n"+
		//	"password: %v\n"+
		//	"database: %v\n"+
		//	"rootPath: %v\n"+
		//	"rootPackagePath: %v\n", host, *port, user, password, databaseName, rootPath, rootPackagePath)
		if "" == host {
			host = interact.AskDBHost()
		}
		if 0 == *port {
			*port = interact.AskDBPort()
		}
		if "" == user {
			user = interact.AskDBUser()
		}
		if "" == password {
			password = interact.AskDBPassword()
		}
		if rootPackagePath == "" {
			rootPackagePath = interact.AskPackage()
		}
		if "" == entityPackage {
			entityPackage = interact.AskEntityPackage()
		}
		if "" == mapperPackage {
			mapperPackage = interact.AskMapperPackage()
		}
		if "" == mapperXmlPath {
			mapperXmlPath = interact.AskMapperXmlPath()
		}
		if "" == queryPackage {
			queryPackage = interact.AskQueryPackage()
		}
		if index := strings.LastIndex(queryPackage, "."); -1 == index {
			queryRootPackage = queryPackage
		} else {
			queryRootPackage = queryPackage[0:index]
		}
		if "" == rootPath {
			rootPath = interact.AskExportPath()
		}

		tableNames = nil
		if nil == args || 0 == len(args) {
			databaseName = ""
		} else {
			if 1 < len(args) {
				tableNames = make([]string, len(args)-1)
			}
			for i, v := range args {
				if 0 == i {
					databaseName = strings.Trim(v, "\"' \t\n")
				} else {
					tableNames[i-1] = strings.Trim(v, "\"' \t\n")
				}
			}
		}
		if databaseName == "" {
			// fmt.Printf("Database name can not be null")
			databaseName = interact.AskDBName()
		}
		if tablePrefix == "" {
			tablePrefix = interact.AskTablePrefix()
		}
		if *allTable {
			tableNames = nil
		} else {
			if nil == tableNames || 0 == len(tableNames) { // 未填写tableNames的情况下
				isAllTable := interact.AskIsAllTableOfDB()
				if isAllTable {
					tableNames = nil
				} else {
					tableNames = interact.AskTables()
				}
			}
		}
		if *overwriteAll {
			conflictOverwriteAll = true
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=1&multiStatements=1&charset=utf8mb4&collation=utf8mb4_unicode_ci", user, password, "tcp", host, *port, "information_schema")
		var err error
		config.DbIns, err = sql.Open("mysql", dsn)
		if nil != err {
			color.Red("Open mysql failed, err: %v\n", err)
			return
		}
		defer config.DbIns.Close()
		//最大连接周期，超过时间的连接就close
		config.DbIns.SetConnMaxLifetime(100 * time.Second)
		//设置最大连接数
		config.DbIns.SetMaxOpenConns(100)
		//设置闲置连接数
		config.DbIns.SetMaxIdleConns(16)

		//if err = config.DbIns.Ping(); nil != err {
		//    return errors.New(fmt.Sprintf("Connect to mysql faild, err: %v", err))
		//}
		// 查询出所有的表
		var rows *sql.Rows
		if nil != tableNames && 0 < len(tableNames) {
			for i, v := range tableNames {
				tableNames[i] = strings.Trim(v, "\"' \t\n")
			}
			// params := make([]string, len(tableNames)+1)
			params := []interface{}{databaseName}

			for _, v := range tableNames {
				params = append(params, v)
			}
			rows, err = config.DbIns.Query("select TABLE_NAME as TableName, TABLE_COMMENT as `Comment` from TABLES where TABLE_SCHEMA = ? and TABLE_NAME in (?"+strings.Repeat(",?", len(tableNames)-1)+")", params...)
			if nil != err {
				fmt.Printf("Query all table of %s failed. err: %v\n", databaseName, err)
				config.DbIns.Close()
				os.Exit(-1)
			}
		} else {
			rows, err = config.DbIns.Query("select TABLE_NAME as TableName, TABLE_COMMENT as `Comment` from TABLES where TABLE_SCHEMA = ?", databaseName)
			if nil != err {
				fmt.Printf("Query all table of %s failed. err: %v\n", databaseName, err)
				config.DbIns.Close()
				os.Exit(-1)
			}
		}
		defer rows.Close()
		// 初始化query
		var templateData TemplateData
		templateData.EntityPackage = entityPackage
		templateData.QueryPackage = queryPackage
		templateData.MapperPackage = mapperPackage
		templateData.QueryRootPackage = queryRootPackage
		templateData.PackagePath = rootPackagePath
		templateData.TableNameHump = "Query"
		if err := generate("", config.BaseQueryTemp, queryRootPackage, "java", &templateData); nil != err {
			color.Red("Generate base query[%s.%s.Query] failed, err: %s\n", rootPackagePath, queryRootPackage, err.Error())
		} else {
			color.Green("Generate base query[%s.%s.Query] success.", rootPackagePath, queryRootPackage)
		}
		for rows.Next() {
			var tableName table
			if err := rows.Scan(&tableName.TableName, &tableName.Comment); nil != err {
				fmt.Printf("Scan rows failed, err: %v\n", err)
				return
			}
			var templateData TemplateData
			templateData.TableName = tableName.TableName
			templateData.EntityPackage = entityPackage
			templateData.QueryPackage = queryPackage
			templateData.MapperPackage = mapperPackage
			templateData.QueryRootPackage = queryRootPackage
			templateData.TableNameHump = toHump(strings.TrimPrefix(tableName.TableName, tablePrefix), true)
			templateData.TableNote = tableName.Comment
			templateData.PackagePath = rootPackagePath
			generateTable(&templateData)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cpm.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().StringP("source", "s", ".", "The source for pro")
	//current, err := USER.Current()
	//if nil != err {
	//	fmt.Printf("Load user info failed, err: %v", err)
	//	os.Exit(-1)
	//}
	//defaultDocumentRoot := fmt.Sprintf("%s%cDocuments%cexports", current.HomeDir, filepath.Separator, filepath.Separator)
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "The host of mysql")
	port = rootCmd.PersistentFlags().Uint16P("port", "P", 0, "The port of mysql")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "The username of mysql")
	rootCmd.PersistentFlags().StringVarP(&entityPackage, "entity-package", "e", "", "The package of the entity that needs to be generated, not including the root package")
	rootCmd.PersistentFlags().StringVarP(&mapperPackage, "mapper-package", "m", "", "The package of the mapper that needs to be generated, not including the root package")
	rootCmd.PersistentFlags().StringVarP(&mapperXmlPath, "mapper-path", "M", "", "The path of the mapper xml that needs to be generated, not including the root package")
	rootCmd.PersistentFlags().StringVarP(&queryPackage, "query-package", "q", "", "The package of the query that needs to be generated, not including the root package")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "the password of mysql")
	rootCmd.PersistentFlags().StringVar(&rootPath, "root-path", "", "the path of export directory")
	rootCmd.PersistentFlags().StringVar(&rootPackagePath, "package", "", "the package path of generate, e.g: \"work.bottle\"")
	rootCmd.PersistentFlags().StringVar(&tablePrefix, "table-prefix", "", "the table prefix of table name")
	overwriteAll = rootCmd.PersistentFlags().BoolP("overwrite", "o", false, "overwrite all of exists files")
	allTable = rootCmd.PersistentFlags().BoolP("all-table", "a", false, "generator all of table")
}

func generateTable(temp *TemplateData) {
	// fmt.Printf("TableName is : %v, TableNameHump: %v, pointer: %p\n", temp.TableName, temp.TableNameHump, &temp)
	rows, err := config.DbIns.Query("select `COLUMN_NAME` as Field, `DATA_TYPE` as DataType, `COLUMN_KEY` as `Index`, `COLUMN_COMMENT` as Comment from `COLUMNS` where TABLE_SCHEMA = ? AND TABLE_NAME = ?", databaseName, temp.TableName)
	if nil != err {
		fmt.Println("Query table %v failed, err: %v", temp.TableName, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var column column
		if err := rows.Scan(&column.Field, &column.DataType, &column.Index, &column.Comment); nil != err {
			fmt.Printf("Scan rows failed, err: %v\n", err)
			return
		}
		column.Property = toHump(column.Field, false)
		column.PropertyN = toHump(column.Field, true)
		if column.Index == "PRI" {
			column.IsPk = 1
			temp.Pk = column.Field
			temp.PkHump = column.Property
		}
		if column.Index == "PRI" || column.Index == "MUL" || column.Index == "UNI" {
			column.IsIndex = 1
		}
		switch column.DataType {
		case "int", "integer", "INT", "INTEGER":
			column.JdbcType = "INTEGER"
			column.JavaType = "Integer"
		case "mediumint", "MEDIUMINT":
			column.JdbcType = "INTEGER"
			column.JavaType = "Integer"
		case "varchar", "VARCHAR":
			column.JdbcType = "VARCHAR"
			column.JavaType = "String"
		case "tinyint", "TINYINT":
			column.JdbcType = "TINYINT"
			column.JavaType = "Integer"
		case "timestamp", "datetime", "TIMESTAMP", "DATETIME":
			column.JdbcType = "TIMESTAMP"
			column.JavaType = "java.sql.Timestamp"
		case "time", "TIME":
			column.JdbcType = "TIME"
			column.JavaType = "java.sql.Time"
		case "smallint", "SMALLINT":
			column.JdbcType = "SMALLINT"
			column.JavaType = "Integer"
		case "real", "REAL":
			column.JdbcType = "REAL"
			column.JavaType = "Object"
		case "numeric", "NUMERIC":
			column.JdbcType = "NUMERIC"
			column.JavaType = "BigDecimal"
		case "float", "FLOAT":
			column.JdbcType = "FLOAT"
			column.JavaType = "Float"
		case "double", "DOUBLE":
			column.JdbcType = "DOUBLE"
			column.JavaType = "Double"
		case "decimal", "DECIMAL":
			column.JdbcType = "DECIMAL"
			column.JavaType = "BigDecimal"
		case "date", "DATE":
			column.JdbcType = "DATE"
			column.JavaType = "java.sql.Date"
		case "clob", "CLOB", "text", "TEXT":
			column.JdbcType = "CLOB"
			column.JavaType = "String"
		case "char", "CHAR":
			column.JdbcType = "CHAR"
			column.JavaType = "String"
		case "blob", "BLOB":
			column.JdbcType = "BLOB"
			column.JavaType = "Byte[]"
		case "bit", "BIT":
			column.JdbcType = "BIT"
			column.JavaType = "Byte"
		case "bigint", "BIGINT":
			column.JdbcType = "BIGINT"
			column.JavaType = "Long"
		default:
			column.JdbcType = ""
			column.JavaType = "Object"

		}
		temp.Fields = append(temp.Fields, column)
		// fmt.Printf("Field: %v, Property: %v, DataType: %v, Index: %v, IsIndex: %v, IsPk: %v, Comment: %v\n", column.Field, column.Property, column.DataType, column.Index, column.IsIndex, column.IsPk, column.Comment)
	}

	if err := generate("", config.EntityTemp, entityPackage, "java", temp); nil != err {
		color.Red("Generate entity[%s.%s.%s] failed, err: %s\n", temp.PackagePath, temp.EntityPackage, temp.TableNameHump, err.Error())
	} else {
		color.Green("Generate entity[%s.%s.%s] success.", temp.PackagePath, temp.EntityPackage, temp.TableNameHump)
	}
	if err := generate("query", config.QueryTemp, queryPackage, "java", temp); nil != err {
		color.Red("Generate query[%s.%s.%sQuery] failed, err: %s\n", temp.PackagePath, temp.QueryPackage, temp.TableNameHump, err.Error())
	} else {
		color.Green("Generate query[%s.%s.%sQuery] success.", temp.PackagePath, temp.QueryPackage, temp.TableNameHump)
	}
	if err := generate("mapper", config.MapperTemp, mapperPackage, "java", temp); nil != err {
		color.Red("Generate mapper[%s.%s.%sMapper] failed, err: %s\n", temp.PackagePath, temp.MapperPackage, temp.TableNameHump, err.Error())
	} else {
		color.Green("Generate mapper[%s.%s.%sMapper] success.", temp.PackagePath, temp.MapperPackage, temp.TableNameHump)
	}
	if err := generate("mapper", config.MapperXmlTemp, mapperXmlPath, "xml", temp); nil != err {
		color.Red("Generate mapper xml[%s%c%s%c%sMapper.xml] failed, err: %s\n", temp.PackagePath, filepath.Separator, mapperXmlPath, filepath.Separator, temp.TableNameHump, err.Error())
	} else {
		color.Green("Generate mapper xml[%s%c%s%c%sMapper.xml] success.", temp.PackagePath, filepath.Separator, mapperXmlPath, filepath.Separator, temp.TableNameHump)
	}
}

func generate(title, tempStr, pkg, suffix string, temp *TemplateData) error {
	var errStr string
	var fPath string
	if "" == pkg {
		fPath = fmt.Sprintf("%s%c%s%s.%s", rootPath, filepath.Separator, temp.TableNameHump, toHump(title, true), suffix)
	} else {
		fPath = fmt.Sprintf("%s%c%s%c%s%s.%s", rootPath, filepath.Separator,
			strings.ReplaceAll(pkg, ".", string(filepath.Separator)), filepath.Separator, temp.TableNameHump, toHump(title, true), suffix)
	}

	stat, err := os.Stat(fPath)
	if nil != err {
		if !os.IsNotExist(err) {
			fmt.Sprintf(errStr, "Failed to generate %s, err: %v", title, err)
			return errors.New(errStr)
		}
	} else {
		if stat.IsDir() {
			fmt.Sprintf(errStr, "The file already exists, but it is a directory[%s]", fPath)
			return errors.New(errStr)
		} else {
			if conflictOverwriteAll {
				// do nothing
			} else {
				isOverwrite := interact.AskIsOverwrite(fPath)
				if "overwrite all" == isOverwrite {
					conflictOverwriteAll = true
				} else if "overwrite" == isOverwrite {
					// do nothing
				} else {
					// do not overwrite
					return nil
				}
			}
		}
	}

	// 生成它的父目录
	dir, _ := filepath.Split(fPath)
	if err = os.MkdirAll(dir, 0750); nil != err {
		fmt.Sprintf(errStr, "Create %s directory failed, err: %v", title, err)
		return errors.New(errStr)
	}
	file, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0750)
	if nil != err {
		fmt.Sprintf(errStr, "Open file[%s] failed, err: %v", fPath, err)
		return errors.New(errStr)
	}
	defer file.Close()
	tempEntity, err := template.New(title).Parse(tempStr) // （2）解析模板
	if err != nil {
		errStr = "template parse failed"
		return errors.New(errStr)
	}
	err = tempEntity.Execute(file, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		errStr = "write to file failed"
		return errors.New(errStr)
	}
	return nil
}

func toHump(source string, first bool) string {
	if "" == source {
		return ""
	}
	split := strings.Split(source, "_")
	for i, s := range split {
		if !first && 0 == i {
			continue
		}
		strArry := []rune(s)
		if strArry[0] >= 97 && strArry[0] <= 122 {
			strArry[0] -= 32
		}
		split[i] = string(strArry)
	}
	return strings.Join(split, "")
}
