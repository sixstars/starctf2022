package libraryelements

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

type Pair struct {
	key   string
	value interface{}
}

func selectLibraryElementByParam(params []Pair) (string, []interface{}) {
	conditions := make([]string, 0, len(params))
	values := make([]interface{}, 0, len(params))
	for _, p := range params {
		conditions = append(conditions, "le."+p.key+"=?")
		values = append(values, p.value)
	}
	return ` WHERE ` + strings.Join(conditions, " AND "), values
}

func writeParamSelectorSQL(builder *sqlstore.SQLBuilder, params ...Pair) {
	if len(params) > 0 {
		conditionString, paramValues := selectLibraryElementByParam(params)
		builder.Write(conditionString, paramValues...)
	}
}

func writePerPageSQL(query searchLibraryElementsQuery, sqlStore *sqlstore.SQLStore, builder *sqlstore.SQLBuilder) {
	if query.perPage != 0 {
		offset := query.perPage * (query.page - 1)
		builder.Write(sqlStore.Dialect.LimitOffset(int64(query.perPage), int64(offset)))
	}
}

func writeKindSQL(query searchLibraryElementsQuery, builder *sqlstore.SQLBuilder) {
	if models.LibraryElementKind(query.kind) == models.PanelElement || models.LibraryElementKind(query.kind) == models.VariableElement {
		builder.Write(" AND le.kind = ?", query.kind)
	}
}

func writeTypeFilterSQL(typeFilter []string, builder *sqlstore.SQLBuilder) {
	if len(typeFilter) > 0 {
		var sql bytes.Buffer
		params := make([]interface{}, 0)
		sql.WriteString(` AND le.type IN (?` + strings.Repeat(",?", len(typeFilter)-1) + ")")
		for _, filter := range typeFilter {
			params = append(params, filter)
		}
		builder.Write(sql.String(), params...)
	}
}

func writeSearchStringSQL(query searchLibraryElementsQuery, sqlStore *sqlstore.SQLStore, builder *sqlstore.SQLBuilder) {
	if len(strings.TrimSpace(query.searchString)) > 0 {
		builder.Write(" AND (le.name "+sqlStore.Dialect.LikeStr()+" ?", "%"+query.searchString+"%")
		builder.Write(" OR le.description "+sqlStore.Dialect.LikeStr()+" ?)", "%"+query.searchString+"%")
	}
}

func writeExcludeSQL(query searchLibraryElementsQuery, builder *sqlstore.SQLBuilder) {
	if len(strings.TrimSpace(query.excludeUID)) > 0 {
		builder.Write(" AND le.uid <> ?", query.excludeUID)
	}
}

type FolderFilter struct {
	includeGeneralFolder bool
	folderIDs            []string
	parseError           error
}

func parseFolderFilter(query searchLibraryElementsQuery) FolderFilter {
	folderIDs := make([]string, 0)
	if len(strings.TrimSpace(query.folderFilter)) == 0 {
		return FolderFilter{
			includeGeneralFolder: true,
			folderIDs:            folderIDs,
			parseError:           nil,
		}
	}

	includeGeneralFolder := false
	folderIDs = strings.Split(query.folderFilter, ",")
	for _, filter := range folderIDs {
		folderID, err := strconv.ParseInt(filter, 10, 64)
		if err != nil {
			return FolderFilter{
				includeGeneralFolder: false,
				folderIDs:            folderIDs,
				parseError:           err,
			}
		}
		if isGeneralFolder(folderID) {
			includeGeneralFolder = true
			break
		}
	}

	return FolderFilter{
		includeGeneralFolder: includeGeneralFolder,
		folderIDs:            folderIDs,
		parseError:           nil,
	}
}

func (f *FolderFilter) writeFolderFilterSQL(includeGeneral bool, builder *sqlstore.SQLBuilder) error {
	var sql bytes.Buffer
	params := make([]interface{}, 0)
	for _, filter := range f.folderIDs {
		folderID, err := strconv.ParseInt(filter, 10, 64)
		if err != nil {
			return err
		}
		if !includeGeneral && isGeneralFolder(folderID) {
			continue
		}
		params = append(params, filter)
	}
	if len(params) > 0 {
		sql.WriteString(` AND le.folder_id IN (?` + strings.Repeat(",?", len(params)-1) + ")")
		builder.Write(sql.String(), params...)
	}

	return nil
}
