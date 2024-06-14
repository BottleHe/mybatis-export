<!-- {{ .TableNote }} -->
<!DOCTYPE mapper
        PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN"
        "http://mybatis.org/dtd/mybatis-3-mapper.dtd">

<mapper namespace="{{.PackagePath}}.{{ .MapperPackage }}.{{ .TableNameHump }}Mapper">
    <resultMap id="{{- .TableNameHump -}}" type="{{- .PackagePath -}}.{{- .EntityPackage -}}.{{- .TableNameHump -}}">
	{{- range $v := .Fields -}}
	{{- if eq $v.IsPk 1 }}
		<id column="{{- $v.Field -}}" property="{{- $v.Property -}}" jdbcType="{{- $v.JdbcType -}}" />
	{{- else }}
	{{- end -}}
	{{- end }}
	{{- range $v := .Fields -}}
    {{- if eq $v.IsPk 1 }}
        <id column="{{- $v.Field -}}" property="{{- $v.Property -}}" jdbcType="{{- $v.JdbcType -}}" />
    {{- else }}
    {{- end -}}
    {{- end }}
    </resultMap>
    <select id="list" resultMap="{{.TableNameHump}}">
        select
        <choose>
            <when test="null != queryFields">
                <foreach collection="queryFields" separator="," item="Field">
                    `${Field}`
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
                and `{{ $v.Field }}` = #{ {{ $v.Property }}, jdbcType={{ $v.JdbcType }} }
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
                {{ .Pk }}
            </otherwise>
        </choose>
        <choose>
            <when test="sortOrder != null">
                ${sortOrder}
            </when>
            <otherwise>
                asc
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
        select count(*) as cnt from {{.TableName}}
        <where>
			{{- range $v := .Fields -}}
			{{ if or (eq $v.IsIndex 1) (eq $v.IsPk 1) }}
            <if test="{{$v.Property}} != null">
                and `{{ $v.Field }}` = #{ {{ $v.Property }}, jdbcType={{ $v.JdbcType }} }
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
                `{{ $v.Field }}` = #{ {{ $v.Property }},jdbcType={{ $v.JdbcType }} },
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
                `{{ $v.Field }}`,
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
