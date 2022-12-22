package config

const (
	BaseQueryTemp = `package {{ .PackagePath }}.{{ .QueryRootPackage }};

import java.io.Serializable;
import java.util.Map;
import java.util.Set;

public abstract class Query<T> implements Serializable {
    private String sortBy;
    private String sortOrder;
    private Integer page;
    private Integer pageCnt;

    private T data;

    private Map<String, String> allowSortBy;
    private Set<String> queryFields;

    public Query() {
        this.page = 1;
        this.pageCnt = 20;
        this.allowSortBy = initAllowSortBy();
        this.queryFields = initQueryFields();
    }

    public String getSortBy() {
        return sortBy;
    }

    public void setSortBy(String sortBy) {
        if (null == allowSortBy) {
            return;
        }
        if (!allowSortBy.containsKey(sortBy)) {
            return;
        }
        this.sortBy = allowSortBy.get(sortBy);
    }

    public String getSortOrder() {
        return sortOrder;
    }

    public void setSortOrder(SortOrder sortOrder) {
        if (null == sortOrder || (sortOrder != SortOrder.ASC && sortOrder != SortOrder.DESC)) {
            this.sortOrder = "DESC";
        } else {
            this.sortOrder = sortOrder.toString();
        }
    }

    public Integer getPage() {
        if (null != this.page && this.page > 0) {
            return this.page;
        }
        return 1;
    }

    public void setPage(Integer page) {
        this.page = page;
    }

    public void nextPage() {
        this.page++;
    }

    public void prevPage() {
        this.page--;
    }

    public Integer getPageCnt() {
        if (null == this.pageCnt || this.pageCnt <= 0) {
            return 20;
        }
        return this.pageCnt;
    }

    public void setPageCnt(Integer pageCnt) {
        this.pageCnt = pageCnt;
    }

    public Integer getOffset() {
        if (null != this.page && this.page > 0) {
            if (null == this.pageCnt || this.pageCnt <= 0) {
                return (this.page - 1) * 20;
            }
            return (this.page - 1) * this.pageCnt;
        } else {
            return 0;
        }
    }

    public Integer getLength() {
        if (null == this.pageCnt || this.pageCnt <= 0) {
            return 20;
        }
        return this.pageCnt;
    }

    protected abstract Map<String, String> initAllowSortBy();
    protected abstract Set<String> initQueryFields();

    public void setAllowSortBy(Map<String, String> allowSortBy) {
        this.allowSortBy = allowSortBy;
    }

    public Map<String, String> getAllowSortBy() {
        return allowSortBy;
    }

    public void setQueryFields(Set<String> queryFields) {
        this.queryFields = queryFields;
    }

    public Set<String> getQueryFields() {
        return queryFields;
    }

    public Set<String> addQueryField(String field) {
        queryFields.add(field);
        return queryFields;
    }

    public Set<String> removeQueryField(String field) {
        queryFields.remove(field);
        return queryFields;
    }

    public T getData() {
        return data;
    }

    public void setData(T data) {
        this.data = data;
    }

    public static enum SortOrder {
        ASC("ASC"),
        DESC("DESC");

        private String value;

        private SortOrder(String value) {
            this.value = value;
        }

        @Override
        public String toString() {
            return value;
        }
    }
}
`
	QueryTemp = `package {{ .PackagePath }}.{{ .QueryPackage }};

import {{ .PackagePath }}.{{ .QueryRootPackage }}.Query;

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
	EntityTemp = `package {{ .PackagePath }}.{{ .EntityPackage }};

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
	MapperTemp = `package {{ .PackagePath }}.{{ .MapperPackage }};

import {{ .PackagePath }}.{{ .EntityPackage }}.{{ .TableNameHump }};
import {{ .PackagePath }}.{{ .QueryPackage }}.{{ .TableNameHump }}Query;
import org.apache.ibatis.annotations.Mapper;
import org.springframework.stereotype.Repository;

import java.util.List;

@Mapper
public interface {{ .TableNameHump }}Mapper {

    public Integer count({{ .TableNameHump }}Query query);

    public List<{{ .TableNameHump }}> list({{ .TableNameHump }}Query query);

    public Integer insert({{ .TableNameHump }} entity);

    public Integer update({{ .TableNameHump }} entity);

	public Integer delete(@Param("{{ .Pk }}") {{ .PkType }} {{ .Pk }});
}
`
	MapperXmlTemp = `<!-- {{ .TableNote }} -->
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<mapper namespace="{{.PackagePath}}.{{ .MapperPackage }}.{{ .TableNameHump }}Mapper">
    <resultMap id="{{- .TableNameHump -}}" type="{{- .PackagePath -}}.{{- .EntityPackage -}}.{{- .TableNameHump -}}">
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
    <update id="update" parameterType="{{.PackagePath}}.{{- .EntityPackage -}}.{{.TableNameHump}}">
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
    <insert id="insert" parameterType="{{ .PackagePath }}.{{ .EntityPackage }}.{{ .TableNameHump }}" keyProperty="{{ .Pk }}" useGeneratedKeys="true">
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
