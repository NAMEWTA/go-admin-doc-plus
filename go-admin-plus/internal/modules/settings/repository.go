package settings

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/platform/database"
)

type repository struct{ dialect database.Dialect }

func (r repository) literal(field string) string {
	if r.dialect == database.DialectPostgres {
		return "strpos(" + field + ", ?) > 0"
	}
	return "instr(" + field + ", ?) > 0"
}
func (r repository) collate(field string) string {
	if r.dialect == database.DialectPostgres {
		return field + ` COLLATE "C"`
	}
	return field + " COLLATE BINARY"
}

func (r repository) listSettings(ctx context.Context, tx database.Tx, category Category, query ListQuery) (SettingPage, error) {
	where, args := " WHERE category = ?", []any{category}
	if query.Search != "" {
		where += " AND (" + r.literal("lower(setting_key)") + " OR " + r.literal("label_key") + ")"
		args = append(args, query.Search, normalizedTextKey(query.Search))
	}
	var page SettingPage
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_values"+where, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,category,setting_key,label,value,description,enabled,revision FROM settings_values"+where+" ORDER BY "+r.collate("label_key")+", id LIMIT ? OFFSET ?", append(args, query.PerPage, (query.Page-1)*query.PerPage)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	page.Rows = []Setting{}
	for rows.Next() {
		var value Setting
		if err := rows.Scan(&value.ID, &value.Category, &value.Key, &value.Label, &value.Value, &value.Description, &value.Enabled, &value.Revision); err != nil {
			return page, err
		}
		page.Rows = append(page.Rows, value)
	}
	return page, rows.Err()
}

func (r repository) createSetting(ctx context.Context, tx database.Tx, value Setting) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO settings_values(id,category,setting_key,label,label_key,value,description,enabled,revision) VALUES(?,?,?,?,?,?,?,?,?)", value.ID, value.Category, value.Key, value.Label, normalizedTextKey(value.Label), value.Value, value.Description, value.Enabled, value.Revision)
	return err
}
func (r repository) updateSetting(ctx context.Context, tx database.Tx, value Setting, revision int64) error {
	result, err := tx.ExecContext(ctx, "UPDATE settings_values SET category=?,setting_key=?,label=?,label_key=?,value=?,description=?,enabled=?,revision=revision+1 WHERE id=? AND revision=?", value.Category, value.Key, value.Label, normalizedTextKey(value.Label), value.Value, value.Description, value.Enabled, value.ID, revision)
	if err != nil {
		return err
	}
	return classifyMutation(ctx, tx, result, "settings_values", value.ID, revision)
}
func (r repository) getSetting(ctx context.Context, tx database.Tx, id string) (Setting, error) {
	var v Setting
	err := tx.QueryRowContext(ctx, "SELECT id,category,setting_key,label,value,description,enabled,revision FROM settings_values WHERE id=?", id).Scan(&v.ID, &v.Category, &v.Key, &v.Label, &v.Value, &v.Description, &v.Enabled, &v.Revision)
	return v, err
}
func (r repository) deleteSetting(ctx context.Context, tx database.Tx, id string, revision int64) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM settings_values WHERE id=? AND revision=?", id, revision)
	if err != nil {
		return err
	}
	return classifyMutation(ctx, tx, result, "settings_values", id, revision)
}

func (r repository) listDictionaries(ctx context.Context, tx database.Tx, query ListQuery) (DictionaryPage, error) {
	where, args := "", []any{}
	if query.Search != "" {
		where = " WHERE " + r.literal("lower(dictionary_key)") + " OR " + r.literal("name_key")
		args = append(args, query.Search, normalizedTextKey(query.Search))
	}
	var page DictionaryPage
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_dictionary_types"+where, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,dictionary_key,name,description,enabled,revision FROM settings_dictionary_types"+where+" ORDER BY "+r.collate("name_key")+", id LIMIT ? OFFSET ?", append(args, query.PerPage, (query.Page-1)*query.PerPage)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	page.Rows = []Dictionary{}
	for rows.Next() {
		var v Dictionary
		if err := rows.Scan(&v.ID, &v.Key, &v.Name, &v.Description, &v.Enabled, &v.Revision); err != nil {
			return page, err
		}
		page.Rows = append(page.Rows, v)
	}
	return page, rows.Err()
}
func (r repository) createDictionary(ctx context.Context, tx database.Tx, v Dictionary) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO settings_dictionary_types(id,dictionary_key,name,name_key,description,enabled,revision) VALUES(?,?,?,?,?,?,?)", v.ID, v.Key, v.Name, normalizedTextKey(v.Name), v.Description, v.Enabled, v.Revision)
	return err
}
func (r repository) getDictionary(ctx context.Context, tx database.Tx, id string) (Dictionary, error) {
	var v Dictionary
	err := tx.QueryRowContext(ctx, "SELECT id,dictionary_key,name,description,enabled,revision FROM settings_dictionary_types WHERE id=?", id).Scan(&v.ID, &v.Key, &v.Name, &v.Description, &v.Enabled, &v.Revision)
	return v, err
}
func (r repository) updateDictionary(ctx context.Context, tx database.Tx, v Dictionary, revision int64) error {
	result, err := tx.ExecContext(ctx, "UPDATE settings_dictionary_types SET dictionary_key=?,name=?,name_key=?,description=?,enabled=?,revision=revision+1 WHERE id=? AND revision=?", v.Key, v.Name, normalizedTextKey(v.Name), v.Description, v.Enabled, v.ID, revision)
	if err != nil {
		return err
	}
	return classifyMutation(ctx, tx, result, "settings_dictionary_types", v.ID, revision)
}
func (r repository) deleteDictionary(ctx context.Context, tx database.Tx, id string, revision int64) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM settings_dictionary_types WHERE id=? AND revision=?", id, revision)
	if err != nil {
		return err
	}
	return classifyMutation(ctx, tx, result, "settings_dictionary_types", id, revision)
}

func (r repository) listItems(ctx context.Context, tx database.Tx, dictionaryID string, query ListQuery) (DictionaryItemPage, error) {
	where, args := " WHERE dictionary_id=?", []any{dictionaryID}
	if query.Search != "" {
		where += " AND (" + r.literal("value_key") + " OR " + r.literal("label_key") + ")"
		args = append(args, normalizedTextKey(query.Search), normalizedTextKey(query.Search))
	}
	var page DictionaryItemPage
	if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_dictionary_items"+where, args...).Scan(&page.Total); err != nil {
		return page, err
	}
	rows, err := tx.QueryContext(ctx, "SELECT id,dictionary_id,item_value,label,sort_order,enabled,revision FROM settings_dictionary_items"+where+" ORDER BY sort_order,"+r.collate("label_key")+",id LIMIT ? OFFSET ?", append(args, query.PerPage, (query.Page-1)*query.PerPage)...)
	if err != nil {
		return page, err
	}
	defer rows.Close()
	page.Rows = []DictionaryItem{}
	for rows.Next() {
		var v DictionaryItem
		if err := rows.Scan(&v.ID, &v.DictionaryID, &v.Value, &v.Label, &v.SortOrder, &v.Enabled, &v.Revision); err != nil {
			return page, err
		}
		page.Rows = append(page.Rows, v)
	}
	return page, rows.Err()
}
func (r repository) createItem(ctx context.Context, tx database.Tx, v DictionaryItem) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO settings_dictionary_items(id,dictionary_id,item_value,value_key,label,label_key,sort_order,enabled,revision) VALUES(?,?,?,?,?,?,?,?,?)", v.ID, v.DictionaryID, v.Value, normalizedTextKey(v.Value), v.Label, normalizedTextKey(v.Label), v.SortOrder, v.Enabled, v.Revision)
	return err
}
func (r repository) getItem(ctx context.Context, tx database.Tx, id string) (DictionaryItem, error) {
	var v DictionaryItem
	err := tx.QueryRowContext(ctx, "SELECT id,dictionary_id,item_value,label,sort_order,enabled,revision FROM settings_dictionary_items WHERE id=?", id).Scan(&v.ID, &v.DictionaryID, &v.Value, &v.Label, &v.SortOrder, &v.Enabled, &v.Revision)
	return v, err
}
func (r repository) updateItem(ctx context.Context, tx database.Tx, v DictionaryItem, revision int64) error {
	result, err := tx.ExecContext(ctx, "UPDATE settings_dictionary_items SET item_value=?,value_key=?,label=?,label_key=?,sort_order=?,enabled=?,revision=revision+1 WHERE id=? AND revision=?", v.Value, normalizedTextKey(v.Value), v.Label, normalizedTextKey(v.Label), v.SortOrder, v.Enabled, v.ID, revision)
	if err != nil {
		return err
	}
	return classifyMutation(ctx, tx, result, "settings_dictionary_items", v.ID, revision)
}
func (r repository) deleteItem(ctx context.Context, tx database.Tx, id string, revision int64) error {
	result, err := tx.ExecContext(ctx, "DELETE FROM settings_dictionary_items WHERE id=? AND revision=?", id, revision)
	if err != nil {
		return err
	}
	return classifyMutation(ctx, tx, result, "settings_dictionary_items", id, revision)
}

func (r repository) options(ctx context.Context, tx database.Tx, key string) ([]DictionaryOption, error) {
	query := "SELECT i.item_value,i.label FROM settings_dictionary_items i JOIN settings_dictionary_types d ON d.id=i.dictionary_id WHERE d.dictionary_key=? AND d.enabled=? AND i.enabled=? ORDER BY i.sort_order," + r.collate("i.label_key") + ",i.id"
	rows, err := tx.QueryContext(ctx, query, key, true, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []DictionaryOption{}
	for rows.Next() {
		var v DictionaryOption
		if err := rows.Scan(&v.Value, &v.Label); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		var exists int
		if err := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM settings_dictionary_types WHERE dictionary_key=? AND enabled=?", key, true).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, ErrNotFound
		}
	}
	return result, nil
}

func classifyMutation(ctx context.Context, tx database.Tx, result sql.Result, table, id string, revision int64) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 1 {
		return nil
	}
	var found int64
	err = tx.QueryRowContext(ctx, "SELECT revision FROM "+table+" WHERE id=?", id).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return ErrConflict
}

func isUnique(dialect database.Dialect, err error) bool {
	if dialect == database.DialectPostgres {
		var state interface{ SQLState() string }
		return errors.As(err, &state) && state.SQLState() == "23505"
	}
	if dialect == database.DialectSQLite {
		var coded interface{ Code() int }
		return errors.As(err, &coded) && coded.Code()&0xff == 19
	}
	return false
}
func isReference(dialect database.Dialect, err error) bool {
	if dialect == database.DialectPostgres {
		var state interface{ SQLState() string }
		return errors.As(err, &state) && state.SQLState() == "23503"
	}
	if dialect == database.DialectSQLite {
		var coded interface{ Code() int }
		return errors.As(err, &coded) && coded.Code()&0xff == 19
	}
	return false
}

var _ = strings.Builder{}
