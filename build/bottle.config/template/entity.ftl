package {{ .PackagePath }}.{{ .EntityPackage }};

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
