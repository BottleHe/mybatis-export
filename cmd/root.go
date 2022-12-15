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
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"mybatis-export/config"
	"mybatis-export/util"
	"os"
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
	tableName    string
	tablePrefix  string

	rootPath    string
	packagePath string

	dbIns    *sql.DB
	interact util.Interact
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
	Pk            string
	PkHump        string
	PackagePath   string
	TableNote     string
	TableName     string
	TableNameHump string
	Fields        []column
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
		//	"packagePath: %v\n", host, *port, user, password, databaseName, rootPath, packagePath)
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
		if databaseName == "" {
			// fmt.Printf("Database name can not be null")
			databaseName = interact.AskDBName()
		}
		if packagePath == "" {
			packagePath = interact.AskPackage()
		}
		if "" == rootPath {
			rootPath = interact.AskExportPath()
		}
		// 连接数据库
		var err error
		dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=1&multiStatements=1&charset=utf8mb4&collation=utf8mb4_unicode_ci", user, password, "tcp", host, *port, "information_schema")
		dbIns, err = sql.Open("mysql", dsn)
		if nil != err {
			return errors.New(fmt.Sprintf("Open mysql failed, err: %v", err))
		}
		//最大连接周期，超过时间的连接就close
		dbIns.SetConnMaxLifetime(100 * time.Second)
		//设置最大连接数
		dbIns.SetMaxOpenConns(100)
		//设置闲置连接数
		dbIns.SetMaxIdleConns(16)

		//if err = dbIns.Ping(); nil != err {
		//    return errors.New(fmt.Sprintf("Connect to mysql faild, err: %v", err))
		//}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// 查询出所有的表
		var rows *sql.Rows
		var err error
		if tableName != "" {
			rows, err = dbIns.Query("select TABLE_NAME as TableName, TABLE_COMMENT as `Comment` from TABLES where TABLE_SCHEMA = ? and TABLE_NAME = ?", databaseName, tableName)
			if nil != err {
				fmt.Printf("Query all table of %s failed. err: %v\n", databaseName, err)
				dbIns.Close()
				os.Exit(-1)
			}
		} else {
			rows, err = dbIns.Query("select TABLE_NAME as TableName, TABLE_COMMENT as `Comment` from TABLES where TABLE_SCHEMA = ?", databaseName)
			if nil != err {
				fmt.Printf("Query all table of %s failed. err: %v\n", databaseName, err)
				dbIns.Close()
				os.Exit(-1)
			}
		}
		defer rows.Close()
		for rows.Next() {
			var tableName table
			if err := rows.Scan(&tableName.TableName, &tableName.Comment); nil != err {
				fmt.Printf("Scan rows failed, err: %v\n", err)
				return
			}
			var templateData TemplateData
			templateData.TableName = tableName.TableName
			templateData.TableNameHump = toHump(strings.TrimPrefix(tableName.TableName, tablePrefix), true)
			templateData.TableNote = tableName.Comment
			templateData.PackagePath = packagePath
			generateTable(templateData)
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
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "the password of mysql")
	rootCmd.PersistentFlags().StringVarP(&databaseName, "database", "D", "", "the database of mysql")
	rootCmd.PersistentFlags().StringVarP(&tableName, "table", "t", "", "the table name of database")
	rootCmd.PersistentFlags().StringVar(&rootPath, "root-path", "", "the path of export directory")
	rootCmd.PersistentFlags().StringVar(&packagePath, "package", "", "the package path of generate, e.g: \"work.bottle\"")
	rootCmd.PersistentFlags().StringVar(&tablePrefix, "table-prefix", "", "the table prefix of table name")
}

func generateTable(temp TemplateData) {
	// fmt.Printf("TableName is : %v, TableNameHump: %v, pointer: %p\n", temp.TableName, temp.TableNameHump, &temp)
	rows, err := dbIns.Query("select `COLUMN_NAME` as Field, `DATA_TYPE` as DataType, `COLUMN_KEY` as `Index`, `COLUMN_COMMENT` as Comment from `COLUMNS` where TABLE_NAME = ?", temp.TableName)
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

	fmt.Printf("\n\n====================================ENTITY[%s]=======================================\n\n", temp.TableNameHump)
	tempEntity, err := template.New("Entity").Parse(config.EntityTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = tempEntity.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n====================================QUERY[%sQuery]=======================================\n\n", temp.TableNameHump)
	tempQuery, err := template.New("Query").Parse(config.QueryTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = tempQuery.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n====================================MAPPER[%sMapper]=======================================\n\n", temp.TableNameHump)
	tempMapper, err := template.New("Mapper").Parse(config.MapperTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = tempMapper.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n====================================MAPPER_XML[%sMapper.xml]=======================================\n\n", temp.TableNameHump)
	templ, err := template.New("MapperXMLFile").Parse(config.MapperXmlTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = templ.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}
}

func toHump(source string, first bool) string {
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
