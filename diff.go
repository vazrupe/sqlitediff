package sqlitediff

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	sqlite "github.com/mattn/go-sqlite3"
)

// DiffDatabase change database info
type DiffDatabase struct {
	ChanageTables map[string]DiffTable
	AddTables     []string
	RemoveTables  []string
}

// DiffTable is one table's difference before and after rows
// RemoveRows is BeforeTable data
type DiffTable struct {
	Name string

	ChangeSchema bool
	BeforeSQL    string
	AfterSQL     string

	ChangeRows []int64
	AddRows    []int64
	RemoveRows []int64
}

var binded = false

// Diff returns Diff datas
func Diff(before, after string) (*DiffDatabase, error) {
	if !binded {
		sql.Register("sqlite3_hash", &sqlite.SQLiteDriver{
			ConnectHook: func(conn *sqlite.SQLiteConn) error {
				if err := conn.RegisterFunc("md5", hash1, true); err != nil {
					return err
				}
				return nil
			},
		})
	}

	var diffDb = &DiffDatabase{}

	beforeDb, err := sql.Open("sqlite3_hash", before)
	if err != nil {
		return nil, err
	}
	afterDb, err := sql.Open("sqlite3_hash", after)
	if err != nil {
		return nil, err
	}

	beforeTablesMap, err := getTables(beforeDb)
	if err != nil {
		return nil, err
	}
	afterTablesMap, err := getTables(afterDb)
	if err != nil {
		return nil, err
	}

	diffDb.ChanageTables = make(map[string]DiffTable)
	for tableName, tableQuery := range afterTablesMap {
		beforeQuery, exist := beforeTablesMap[tableName]
		// Add Table Case
		if !exist {
			diffDb.AddTables = append(diffDb.AddTables, tableName)
			continue
		}
		delete(beforeTablesMap, tableName)
		// Change Table Case
		var diffTb = DiffTable{Name: tableName,
			ChangeSchema: beforeQuery != tableQuery,
			BeforeSQL:    beforeQuery,
			AfterSQL:     tableQuery}

		beforeKeys := getPrimaryKeys(beforeQuery)
		afterKeys := getPrimaryKeys(tableQuery)

		if len(beforeKeys) == 0 || len(afterKeys) == 0 {
			// Primary Key Not Exist Case
			beforeRows, err := getRowHashs(tableName, beforeDb)
			if err != nil {
				return nil, err
			}
			afterRows, err := getRowHashs(tableName, afterDb)
			if err != nil {
				return nil, err
			}

			for hash, rowid := range afterRows {
				_, beforeExistRow := beforeRows[hash]
				// Add Row Case
				delete(beforeRows, hash)
				if !beforeExistRow {
					diffTb.AddRows = append(diffTb.AddRows, rowid)
					continue
				}
				// No Case: Change Row
			}
			// Remove Row Case
			for _, rowid := range beforeRows {
				diffTb.RemoveRows = append(diffTb.AddRows, rowid)
			}
		} else {
			// Primary Key Exist Case
			beforeRows, err := getKeyRowHashs(tableName, beforeKeys, beforeDb)
			if err != nil {
				return nil, err
			}
			afterRows, err := getKeyRowHashs(tableName, afterKeys, afterDb)
			if err != nil {
				return nil, err
			}

			for key, item := range afterRows {
				beforeRow, beforeExistRow := beforeRows[key]
				// Add Row Case
				if !beforeExistRow {
					diffTb.AddRows = append(diffTb.AddRows, item.ID)
					continue
				}
				delete(beforeRows, key)
				// Change Row Case
				if item.RowHash != beforeRow.RowHash {
					diffTb.ChangeRows = append(diffTb.AddRows, item.ID)
				}
			}
			// Remove Row Case
			for _, item := range beforeRows {
				diffTb.RemoveRows = append(diffTb.AddRows, item.ID)
			}
		}

		if diffTb.ChangeSchema || len(diffTb.ChangeRows) > 0 ||
			len(diffTb.AddRows) > 0 || len(diffTb.RemoveRows) > 0 {
			diffDb.ChanageTables[diffTb.Name] = diffTb
		}
	}
	// Remove Table Case
	for tableName := range beforeTablesMap {
		diffDb.RemoveTables = append(diffDb.RemoveTables, tableName)
	}

	return diffDb, err
}

func getTables(db *sql.DB) (map[string]string, error) {
	rows, err := db.Query("SELECT name, sql FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tableMap := make(map[string]string)
	for rows.Next() {
		var tableName, tableSQL string
		err := rows.Scan(&tableName, &tableSQL)
		if err != nil {
			return tableMap, err
		}
		tableMap[tableName] = tableSQL
	}
	return tableMap, nil
}

func getPrimaryKeys(query string) []string {
	findPrimary := regexp.MustCompile("(?i)primary\\s*key\\s*\\((.+?)\\)")
	keys := regexp.MustCompile("([[:word:]]+)")

	primary := findPrimary.FindAllStringSubmatch(query, 1)
	if len(primary) == 0 {
		return []string{}
	}
	return keys.FindAllString(primary[0][1], -1)
}

func getRowHashs(tableName string, db *sql.DB) (map[string]int64, error) {
	query := fmt.Sprintf("SELECT rowid, md5(*) FROM %s", tableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rowHashs := make(map[string]int64)
	for rows.Next() {
		var id int64
		var hash string
		err := rows.Scan(&id, &hash)
		if err != nil {
			return rowHashs, err
		}
		rowHashs[hash] = id
	}
	return rowHashs, nil
}

type hashrow struct {
	ID      int64
	KeyHash string
	RowHash string
}

func getKeyRowHashs(tableName string, primaryKeys []string, db *sql.DB) (map[string]hashrow, error) {
	query := fmt.Sprintf("SELECT rowid, md5(%s), md5(*) FROM %s",
		keysQueryBuiltin(primaryKeys),
		tableName)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rowHashs := make(map[string]hashrow)
	for rows.Next() {
		var row hashrow
		err := rows.Scan(&row.ID, &row.KeyHash, &row.RowHash)
		if err != nil {
			return rowHashs, err
		}
		rowHashs[row.KeyHash] = row
	}
	return rowHashs, nil
}

func keysQueryBuiltin(keys []string) string {
	return "`" + strings.Join(keys[:], "`,`") + "`"
}
