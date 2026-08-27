package generator

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	modulePattern   = regexp.MustCompile(`^[a-z][a-z0-9]{1,31}$`)
	pathWordPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,63}$`)
	goNamePattern   = regexp.MustCompile(`^[A-Z][A-Za-z0-9]{0,63}$`)
)

func normalize(table Table, draft Draft) (Model, error) {
	if table.Ref != draft.Table || !modulePattern.MatchString(draft.Module) || !goNamePattern.MatchString(draft.Entity) || !pathWordPattern.MatchString(draft.Plural) {
		return Model{}, ErrInvalid
	}
	metadata := make(map[string]Column, len(table.Columns))
	for _, column := range table.Columns {
		if !databaseIdentifierPattern.MatchString(column.Name) || column.Ordinal < 1 {
			return Model{}, ErrInvalid
		}
		if _, duplicate := metadata[column.Name]; duplicate {
			return Model{}, ErrInvalid
		}
		metadata[column.Name] = column
	}
	seenNames, seenFields := map[string]struct{}{}, map[string]struct{}{}
	columns := make([]NormalizedColumn, 0, len(draft.Columns))
	primaryKeys := 0
	for _, configured := range draft.Columns {
		column, exists := metadata[configured.Name]
		if !exists || !goNamePattern.MatchString(configured.Field) {
			return Model{}, ErrInvalid
		}
		if _, duplicate := seenNames[configured.Name]; duplicate {
			return Model{}, ErrInvalid
		}
		if _, duplicate := seenFields[configured.Field]; duplicate {
			return Model{}, ErrInvalid
		}
		seenNames[configured.Name], seenFields[configured.Field] = struct{}{}, struct{}{}
		if !configured.Include {
			if column.PrimaryKey || configured.Searchable || configured.Sortable {
				return Model{}, ErrInvalid
			}
			continue
		}
		normalized, ok := normalizeColumn(column, configured)
		if !ok {
			return Model{}, ErrInvalid
		}
		if normalized.PrimaryKey {
			primaryKeys++
		}
		columns = append(columns, normalized)
	}
	if len(seenNames) != len(metadata) || len(columns) < 5 || primaryKeys != 1 {
		return Model{}, ErrInvalid
	}
	sort.Slice(columns, func(i, j int) bool { return columns[i].Ordinal < columns[j].Ordinal })
	model := Model{Module: draft.Module, Entity: draft.Entity, EntityVariable: lowerFirst(draft.Entity), Plural: draft.Plural, TableName: table.Ref.Name, Schema: table.Ref.Schema, Columns: columns}
	for _, column := range columns {
		if column.PrimaryKey {
			model.PrimaryKey = column
			break
		}
	}
	standard := map[string]ColumnKind{"id": KindUUID, "revision": KindInt64, "created_at": KindTime, "updated_at": KindTime}
	for name, kind := range standard {
		found := false
		for _, column := range columns {
			if column.Name == name && column.Kind == kind && !column.Nullable {
				found = true
			}
		}
		if !found {
			return Model{}, ErrInvalid
		}
	}
	if model.PrimaryKey.Name != "id" || model.PrimaryKey.Field != "ID" {
		return Model{}, ErrInvalid
	}
	return model, nil
}

func normalizeColumn(column Column, draft ColumnDraft) (NormalizedColumn, bool) {
	value := NormalizedColumn{Name: column.Name, Field: draft.Field, DatabaseType: column.DatabaseType, Kind: column.Kind, Nullable: column.Nullable, PrimaryKey: column.PrimaryKey, Searchable: draft.Searchable, Sortable: draft.Sortable, Ordinal: column.Ordinal}
	switch column.Kind {
	case KindString, KindUUID:
		value.GoType, value.TypeScriptType, value.OpenAPIType = "string", "string", "string"
		if column.Kind == KindUUID {
			value.OpenAPIFormat = "uuid"
		}
	case KindInt64:
		value.GoType, value.TypeScriptType, value.OpenAPIType, value.OpenAPIFormat = "int64", "number", "integer", "int64"
	case KindBoolean:
		value.GoType, value.TypeScriptType, value.OpenAPIType = "bool", "boolean", "boolean"
	case KindDecimal:
		value.GoType, value.TypeScriptType, value.OpenAPIType, value.OpenAPIFormat = "float64", "number", "number", "double"
	case KindTime:
		value.GoType, value.TypeScriptType, value.OpenAPIType, value.OpenAPIFormat = "time.Time", "string", "string", "date-time"
	case KindBytes:
		value.GoType, value.TypeScriptType, value.OpenAPIType, value.OpenAPIFormat = "[]byte", "string", "string", "byte"
	default:
		return NormalizedColumn{}, false
	}
	if column.Nullable {
		value.GoType = "*" + value.GoType
		value.TypeScriptType += " | null"
	}
	return value, true
}

func lowerFirst(value string) string {
	runes := []rune(value)
	for index := 0; index < len(runes); index++ {
		if !unicode.IsUpper(runes[index]) {
			break
		}
		if index > 0 && index+1 < len(runes) && unicode.IsLower(runes[index+1]) {
			break
		}
		runes[index] = unicode.ToLower(runes[index])
	}
	return string(runes)
}

func snakeFromField(value string) string {
	var result strings.Builder
	for index, character := range value {
		if unicode.IsUpper(character) && index > 0 {
			result.WriteByte('_')
		}
		result.WriteRune(unicode.ToLower(character))
	}
	return result.String()
}
