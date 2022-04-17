package migrator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana/pkg/util/errutil"
	"github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

type SQLite3 struct {
	BaseDialect
}

func NewSQLite3Dialect(engine *xorm.Engine) Dialect {
	d := SQLite3{}
	d.BaseDialect.dialect = &d
	d.BaseDialect.engine = engine
	d.BaseDialect.driverName = SQLite
	return &d
}

func (db *SQLite3) SupportEngine() bool {
	return false
}

func (db *SQLite3) Quote(name string) string {
	return "`" + name + "`"
}

func (db *SQLite3) AutoIncrStr() string {
	return "AUTOINCREMENT"
}

func (db *SQLite3) BooleanStr(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func (db *SQLite3) DateTimeFunc(value string) string {
	return "datetime(" + value + ")"
}

func (db *SQLite3) SQLType(c *Column) string {
	switch c.Type {
	case DB_Date, DB_DateTime, DB_TimeStamp, DB_Time:
		return DB_DateTime
	case DB_TimeStampz:
		return DB_Text
	case DB_Char, DB_Varchar, DB_NVarchar, DB_TinyText, DB_Text, DB_MediumText, DB_LongText:
		return DB_Text
	case DB_Bit, DB_TinyInt, DB_SmallInt, DB_MediumInt, DB_Int, DB_Integer, DB_BigInt, DB_Bool:
		return DB_Integer
	case DB_Float, DB_Double, DB_Real:
		return DB_Real
	case DB_Decimal, DB_Numeric:
		return DB_Numeric
	case DB_TinyBlob, DB_Blob, DB_MediumBlob, DB_LongBlob, DB_Bytea, DB_Binary, DB_VarBinary:
		return DB_Blob
	case DB_Serial, DB_BigSerial:
		c.IsPrimaryKey = true
		c.IsAutoIncrement = true
		c.Nullable = false
		return DB_Integer
	default:
		return c.Type
	}
}

func (db *SQLite3) IndexCheckSQL(tableName, indexName string) (string, []interface{}) {
	args := []interface{}{tableName, indexName}
	sql := "SELECT 1 FROM " + db.Quote("sqlite_master") + " WHERE " + db.Quote("type") + "='index' AND " + db.Quote("tbl_name") + "=? AND " + db.Quote("name") + "=?"
	return sql, args
}

func (db *SQLite3) DropIndexSQL(tableName string, index *Index) string {
	quote := db.Quote
	// var unique string
	idxName := index.XName(tableName)
	return fmt.Sprintf("DROP INDEX %v", quote(idxName))
}

func (db *SQLite3) CleanDB() error {
	return nil
}

// TruncateDBTables deletes all data from all the tables and resets the sequences.
// A special case is the dashboard_acl table where we keep the default permissions.
func (db *SQLite3) TruncateDBTables() error {
	tables, err := db.engine.DBMetas()
	if err != nil {
		return err
	}

	sess := db.engine.NewSession()
	defer sess.Close()

	for _, table := range tables {
		switch table.Name {
		case "migration_log":
			continue
		case "dashboard_acl":
			// keep default dashboard permissions
			if _, err := sess.Exec(fmt.Sprintf("DELETE FROM %q WHERE dashboard_id != -1 AND org_id != -1;", table.Name)); err != nil {
				return errutil.Wrapf(err, "failed to truncate table %q", table.Name)
			}
			if _, err := sess.Exec("UPDATE sqlite_sequence SET seq = 2 WHERE name = '%s';", table.Name); err != nil {
				return errutil.Wrapf(err, "failed to cleanup sqlite_sequence")
			}
		default:
			if _, err := sess.Exec(fmt.Sprintf("DELETE FROM %s;", table.Name)); err != nil {
				return errutil.Wrapf(err, "failed to truncate table %q", table.Name)
			}
		}
	}
	if _, err := sess.Exec("UPDATE sqlite_sequence SET seq = 0 WHERE name != 'dashboard_acl';"); err != nil {
		return errutil.Wrapf(err, "failed to cleanup sqlite_sequence")
	}
	return nil
}

func (db *SQLite3) isThisError(err error, errcode int) bool {
	var driverErr sqlite3.Error
	if errors.As(err, &driverErr) {
		if int(driverErr.ExtendedCode) == errcode {
			return true
		}
	}

	return false
}

func (db *SQLite3) ErrorMessage(err error) string {
	var driverErr sqlite3.Error
	if errors.As(err, &driverErr) {
		return driverErr.Error()
	}
	return ""
}

func (db *SQLite3) IsUniqueConstraintViolation(err error) bool {
	return db.isThisError(err, int(sqlite3.ErrConstraintUnique))
}

func (db *SQLite3) IsDeadlock(err error) bool {
	return false // No deadlock
}

// UpsertSQL returns the upsert sql statement for SQLite dialect
func (db *SQLite3) UpsertSQL(tableName string, keyCols, updateCols []string) string {
	columnsStr := strings.Builder{}
	onConflictStr := strings.Builder{}
	colPlaceHoldersStr := strings.Builder{}
	setStr := strings.Builder{}

	const separator = ", "
	separatorVar := separator
	for i, c := range updateCols {
		if i == len(updateCols)-1 {
			separatorVar = ""
		}

		columnsStr.WriteString(fmt.Sprintf("%s%s", db.Quote(c), separatorVar))
		colPlaceHoldersStr.WriteString(fmt.Sprintf("?%s", separatorVar))
		setStr.WriteString(fmt.Sprintf("%s=excluded.%s%s", db.Quote(c), db.Quote(c), separatorVar))
	}

	separatorVar = separator
	for i, c := range keyCols {
		if i == len(keyCols)-1 {
			separatorVar = ""
		}
		onConflictStr.WriteString(fmt.Sprintf("%s%s", db.Quote(c), separatorVar))
	}

	s := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT(%s) DO UPDATE SET %s`,
		tableName,
		columnsStr.String(),
		colPlaceHoldersStr.String(),
		onConflictStr.String(),
		setStr.String(),
	)
	return s
}
