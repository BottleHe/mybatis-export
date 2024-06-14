package {{ .PackagePath }}.{{ .MapperPackage }};

import {{ .PackagePath }}.{{ .EntityPackage }}.{{ .TableNameHump }};
import {{ .PackagePath }}.{{ .QueryPackage }}.{{ .TableNameHump }}Query;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

@Mapper
public interface {{ .TableNameHump }}Mapper {

    public Integer count({{ .TableNameHump }}Query query);

    public List<{{ .TableNameHump }}> list({{ .TableNameHump }}Query query);

    public Integer insert({{ .TableNameHump }} entity);

    public Integer update({{ .TableNameHump }} entity);

	public Integer delete(@Param("{{ .Pk }}") {{ .PkType }} {{ .Pk }});
}
