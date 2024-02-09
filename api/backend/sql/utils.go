package sql

import (
	"encoding/json"
	"strconv"
	"strings"
)

func CleanSQL(sql string) string {
	var rows []string
	for _, row := range strings.Split(sql, "\n") {
		row = strings.TrimSpace(row)
		if row != "" {
			rows = append(rows, row)
		}
	}
	return strings.Join(rows, "\n")
}

//
// SqlArgs
//

type SqlArgs struct {
	Args []any
}

func (self *SqlArgs) Add(arg any) string {
	self.Args = append(self.Args, arg)
	return "$" + strconv.Itoa(len(self.Args))
}

//
// SqlWhere
//

type SqlWhere struct {
	where []string
}

func (self *SqlWhere) Add(e string) {
	self.where = append(self.where, e)
}

func (self *SqlWhere) Apply(sql string) string {
	if len(self.where) > 0 {
		return insertSql(sql, "GROUP BY", "WHERE "+strings.Join(self.where, " AND "))
	} else {
		return sql
	}
}

//
// SqlWith
//

type SqlWith struct {
	withs []string
	joins []string
}

func (self *SqlWith) Add(table string, id string, select_ string) {
	withTable := "with" + strconv.Itoa(len(self.withs))
	self.withs = append(self.withs, withTable+" AS ("+select_+")")
	self.joins = append(self.joins, "JOIN "+withTable+" ON "+table+"."+id+" = "+withTable+"."+id)
}

func (self *SqlWith) Apply(sql string) string {
	if len(self.withs) > 0 {
		sql = "WITH\n" + strings.Join(self.withs, ",\n") + "\n" + sql
		sql = insertSql(sql, "GROUP BY", strings.Join(self.joins, "\n"))
	}
	return sql
}

// Utils

func insertSql(sql string, place string, insert string) string {
	if insert == "" {
		return sql
	} else if place == "" {
		return sql + "\n" + insert
	} else {
		if p := strings.Index(sql, place); p == -1 {
			return sql + "\n" + insert
		} else if p == len(sql)-1 {
			return sql[:p] + insert
		} else {
			return sql[:p] + insert + "\n" + sql[p:]
		}
	}
}

func jsonUnmarshallMapEntries(mapJson []byte, map_ map[string]string) error {
	if len(mapJson) == 0 {
		return nil
	}

	var metadata_ [][]string
	if err := json.Unmarshal(mapJson, &metadata_); err == nil {
		for _, entry := range metadata_ {
			if len(entry) == 2 {
				map_[entry[0]] = entry[1]
			}
		}
		return nil
	} else {
		return err
	}
}

func jsonUnmarshallMap(mapJson []byte, map_ map[string]string) error {
	if err := json.Unmarshal(mapJson, &map_); err == nil {
		return nil
	} else {
		return err
	}
}

func jsonUnmarshallArray(arrayJson []byte, array *[]string) error {
	if len(arrayJson) > 0 {
		return json.Unmarshal(arrayJson, array)
	}
	return nil
}
