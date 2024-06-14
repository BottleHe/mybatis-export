package {{ .PackagePath }}.{{ .QueryPackage }};

import java.io.Serializable;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Set;

public class {{ .TableNameHump }}Query implements Serializable {
	private String sortBy;
    private String sortOrder;
    private Integer page;
    private Integer pageCnt;

	private Map<String, String> allowSortBy;
    private Set<String> queryFields;

    {{- range $v := .Fields -}}
    {{ if eq $v.IsIndex 1 }}
    // {{ $v.Comment }}
    private {{ $v.JavaType }} {{ $v.Property }};
    {{- end -}}
    {{- end }}
    
	public {{ .TableNameHump }}Query() {
        this.page = 1;
        this.pageCnt = 20;
        this.allowSortBy = initAllowSortBy();
        this.queryFields = initQueryFields();
    }

    protected Map<String, String> initAllowSortBy() {
        HashMap<String, String> allowSortByMap = new HashMap<>();
        allowSortByMap.put("{{ .Pk }}", "{{ .Pk }}");
        return allowSortByMap;
    }

    protected Set<String> initQueryFields() {
        HashSet<String> fieldSet = new HashSet<>();
        
        {{ range $v := .Fields -}}
        fieldSet.add("{{ $v.Field }}");
        {{ end }}
        return fieldSet;
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

    public void setSortOrder(String sortOrder) {
		if (null == sortOrder || (!"ASC".equals(sortOrder) && !"DESC".equals(sortOrder))) {
			this.sortOrder = "DESC";
		} else {
			this.sortOrder = sortOrder;
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
