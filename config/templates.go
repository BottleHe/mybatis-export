package config

const (
	QueryTemp = `
package {{ .PackagePath }}.model.query;

import {{ .PackagePath }}.model.Query;

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
	EntityTemp = `
package {{ .PackagePath }}.entity;

import java.io.Serializable;

public class {{ .TableNameHump }} implements Serializable {
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
	MapperTemp = `
package {{ .PackagePath }}.mapper;

import {{ .PackagePath }}.entity.{{ .TableNameHump }};
import {{ .PackagePath }}.model.query.{{ .TableNameHump }}Query;
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
	MapperXmlTemp = `
<!-- {{ .TableNote }} -->
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<mapper namespace="{{.PackagePath}}.mapper.{{ .TableNameHump }}Mapper">
    <resultMap id="{{- .TableNameHump -}}" type="{{- .PackagePath -}}.entity.{{- .TableNameHump -}}">
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
    <update id="update" parameterType="{{.PackagePath}}.entity.{{.TableNameHump}}">
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
    <insert id="insert" parameterType="{{ .PackagePath }}.entity.{{ .TableNameHump }}" keyProperty="{{ .Pk }}" useGeneratedKeys="true">
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
