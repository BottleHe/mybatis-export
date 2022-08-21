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
	"os"
	USER "os/user"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

const (
	queryTemp = `
package {{ .PackagePath }}.entity.dto.query;

import {{ .PackagePath }}.entity.Query;

import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

public class {{ .TableNameHump }}Query extends Query {
    {{- range $v := .Fields -}}
    {{ if eq $v.IsIndex 1 }}
    // {{ $v.Comment }}
    private {{ $v.JavaType }} {{ $v.Property }};
    {{- end -}}
    {{- end }}
    

    @Override
    protected Map<String, String> initAllowSortBy() {
        HashMap<String, String> stringStringHashMap = new HashMap<>();
        stringStringHashMap.put("{{ .Pk }}", "{{ .Pk }}");
        return stringStringHashMap;
    }

    @Override
    protected Set<String> initQueryFields() {
        HashSet<String> fieldSet = new HashSet<>();
        
        {{ range $v := .Fields -}}
        fieldSet.add("{{ $v.Field }}");
        {{ end }}
        return fieldSet;
    }

    {{- range $v := .Fields -}}
    {{- if eq $v.IsIndex 1 }}
    public void set{{- $v.PropertyN }}({{$v.JavaType}} {{$v.Property}}) {
        this.{{$v.Property}} = {{$v.Property}};
    }
    {{- if eq $v.JavaType "Boolean" -}}
    public {{$v.JavaType}} is{{- $v.PropertyN}}() {
        return this.{{$v.Property}};
    }
    {{ else }}
    public {{$v.JavaType}} get{{- $v.PropertyN}}() {
        return this.{{$v.Property}};
    }
    {{- end -}}
    {{- end -}}
    {{- end }}
}
`
	entityTemp = `
package {{ .PackagePath }}.entity.domain;

public class {{ .TableNameHump }} {
    {{ range $v := .Fields }}
    // {{ $v.Comment }}
    private {{ $v.JavaType }} {{ $v.Property }};
    {{ end }}

    {{- range $v := .Fields }}

    public void set{{- $v.PropertyN }}({{$v.JavaType}} {{$v.Property}}) {
        this.{{$v.Property}} = {{$v.Property}};
    }
    {{ if eq $v.JavaType "Boolean" }}
    public {{$v.JavaType}} is{{- $v.PropertyN}}() {
        return this.{{$v.Property}};
    }
    {{ else }}
    public {{$v.JavaType}} get{{- $v.PropertyN}}() {
        return this.{{$v.Property}};
    }
    {{- end -}}
    {{- end }}
}
`
	mapperTemp = `
package {{ .PackagePath }}.mapper;

import {{ .PackagePath }}.entity.domain.{{ .TableNameHump }};
import {{ .PackagePath }}.entity.dto.query.{{ .TableNameHump }}Query;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

import java.util.List;

@Mapper
public interface {{ .TableNameHump }}Mapper {

    public Integer count({{ .TableNameHump }}Query query);

    public List<{{ .TableNameHump }}> list({{ .TableNameHump }}Query query);

    public Integer insert({{ .TableNameHump }} entity);

    public Integer update({{ .TableNameHump }} entity);
}
`
	mapperXmlTemp = `
<!-- {{ .TableNote }} -->
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<mapper namespace="{{.PackagePath}}.mapper.{{ .TableNameHump }}Mapper">
    <resultMap id="{{- .TableNameHump -}}" type="{{- .PackagePath -}}.entity.domain.{{- .TableNameHump -}}">
	{{- range $v := .Fields -}}
	{{- if eq $v.IsPk 1 }}
	<id column="{{- $v.Field -}}" property="{{- $v.Property -}}" jdbcType="{{- $v.JdbcType -}}" />
	{{- else }}
	<result column="{{ $v.Field }}" property="{{ $v.Property }}" jdbcType="{{ $v.JdbcType }}" />
	{{- end -}}
	{{- end }}
    </resultMap>
    <select id="list" resultMap="{{.TableNameHump}}">
        select
        <choose>
            <when test="null != queryFields">
                <foreach collection="queryFields" separator="," item="Field">
                    ` + "`${Field}`" + `
                </foreach>
            </when>
            <otherwise>
                *
            </otherwise>
        </choose>
        from {{ .TableName }}
        <where>
			{{- range $v := .Fields -}}
			{{ if or (eq $v.IsIndex 1) (eq $v.IsPk 1) }}
            <if test="{{$v.Property}} != null">
                and ` + "`{{ $v.Field }}`" + ` = #{ {{ $v.Property }}, jdbcType={{ $v.JdbcType }} }
            </if>
			{{- end }}
			{{- end }}
        </where>
        order by
        <choose>
            <when test="sortBy != null">
                ${sortBy}
            </when>
            <otherwise>
                id
            </otherwise>
        </choose>
        <choose>
            <when test="sortOrder != null">
                ${sortOrder}
            </when>
            <otherwise>
                desc
            </otherwise>
        </choose>
        limit
        <choose>
            <when test="offset != null and offset >= 0">
                #{offset}
            </when>
            <otherwise>
                0
            </otherwise>
        </choose>
        ,
        <choose>
            <when test="length != null and length > 0">
                #{length}
            </when>
            <otherwise>
                20
            </otherwise>
        </choose>
    </select>
    <select id="count" resultType="java.lang.Integer">
        select count(id) as cnt from {{.TableName}}
        <where>
			{{- range $v := .Fields -}}
			{{ if or (eq $v.IsIndex 1) (eq $v.IsPk 1) }}
            <if test="{{$v.Property}} != null">
                and ` + "`{{ $v.Field }}`" + ` = #{ {{ $v.Property }}, jdbcType={{ $v.JdbcType }} }
            </if>
			{{- end }}
			{{- end }}
        </where>
        limit 1
    </select>
    <update id="update" parameterType="{{.PackagePath}}.entity.domain.{{.TableNameHump}}">
        update {{ .TableName }}
        <set>
            {{- range $v := .Fields -}}
            {{ if ne $v.IsPk 1 }}
            <if test="{{ $v.Property }} != null">
                ` + "`{{ $v.Field }}`" + ` = #{ {{ $v.Property }},jdbcType={{ $v.JdbcType }} },
            </if>
			{{- end }}
			{{- end }}
        </set>
        where {{ .Pk }} = #{ {{ .PkHump }} }
    </update>
    <insert id="insert" parameterType="{{ .PackagePath }}.entity.domain.{{ .TableNameHump }}" keyProperty="{{ .Pk }}" useGeneratedKeys="true">
        insert into {{ .TableName }}
        <trim prefix="(" suffix=")" suffixOverrides=",">
            {{- range $v := .Fields -}}
            {{ if eq 1 1 }}
            <if test="{{ $v.Property }} != null">
                ` + "`{{ $v.Field }}`" + `,
            </if>
            {{- end }}
            {{- end }}
        </trim>
        <trim prefix="values(" suffix=")" suffixOverrides=",">
            {{- range $v := .Fields -}}
            {{ if eq 1 1 }}
            <if test="{{ $v.Property }} != null">
                #{ {{ $v.Property }},jdbcType={{ $v.JdbcType }} },
            </if>
            {{- end }}
            {{- end }}
        </trim>
    </insert>
    <delete id="delete">
        delete from {{ .TableName }} where {{ .Pk }} = #{ {{ .PkHump }} }
    </delete>
</mapper>
`
)

var (
	host         string
	user         string
	password     string
	port         *int16
	databaseName string
	tableName    string
	tablePrefix  string

	rootPath    string
	packagePath string

	dbIns *sql.DB
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
		if databaseName == "" {
			// fmt.Printf("Database name can not be null")
			return errors.New("Database name can not be null")
		}
		if packagePath == "" {
			return errors.New("packagePath can not be null")
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

		if err = dbIns.Ping(); nil != err {
			return errors.New(fmt.Sprintf("Connect to mysql faild, err: %v", err))
		}
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
				os.Exit(-1)
			}
		} else {
			rows, err = dbIns.Query("select TABLE_NAME as TableName, TABLE_COMMENT as `Comment` from TABLES where TABLE_SCHEMA = ?", databaseName)
			if nil != err {
				fmt.Printf("Query all table of %s failed. err: %v\n", databaseName, err)
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
	current, err := USER.Current()
	if nil != err {
		fmt.Printf("Load user info failed, err: %v", err)
		os.Exit(-1)
	}
	defaultDocumentRoot := fmt.Sprintf("%s%cDocuments%cexports", current.HomeDir, filepath.Separator, filepath.Separator)
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "localhost", "the host of mysql, default is \"localhost\"")
	port = rootCmd.PersistentFlags().Int16P("port", "P", 3306, "the port of mysql, default is \"3306\"")
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "root", "the username of mysql, default is \"root\"")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "the password of mysql, default is \"\"")
	rootCmd.PersistentFlags().StringVarP(&databaseName, "database", "D", "", "the database of mysql")
	rootCmd.PersistentFlags().StringVarP(&tableName, "table", "t", "", "the table name of database")
	rootCmd.PersistentFlags().StringVar(&rootPath, "root-path", defaultDocumentRoot, "the path of document root, default is \""+defaultDocumentRoot+"\"")
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
	tempEntity, err := template.New("Entity").Parse(entityTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = tempEntity.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n====================================QUERY[%sQuery]=======================================\n\n", temp.TableNameHump)
	tempQuery, err := template.New("Query").Parse(queryTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = tempQuery.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n====================================MAPPER[%sMapper]=======================================\n\n", temp.TableNameHump)
	tempMapper, err := template.New("Mapper").Parse(mapperTemp) // （2）解析模板
	if err != nil {
		panic(err)
	}
	err = tempMapper.Execute(os.Stdout, temp) //（3）数据驱动模板，将name的值填充到模板中
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n====================================MAPPER_XML[%sMapper.xml]=======================================\n\n", temp.TableNameHump)
	templ, err := template.New("MapperXMLFile").Parse(mapperXmlTemp) // （2）解析模板
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
