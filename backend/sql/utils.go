package sql

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/nephio-experimental/tko/util"
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
		clauses := make([]string, len(self.where))
		for index, where := range self.where {
			clauses[index] = "(" + where + ")"
		}

		return insertSqlBefpre(sql, "GROUP BY", "WHERE "+strings.Join(clauses, " AND "))
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

func (self *SqlWith) Add(select_ string, table string, ids ...string) {
	withTable := "with" + strconv.Itoa(len(self.withs))
	self.withs = append(self.withs, withTable+" AS ("+select_+")")

	ons := make([]string, len(ids))
	for index, id := range ids {
		ons[index] = "(" + table + "." + id + " = " + withTable + "." + id + ")"
	}

	/*
		var ons []string
		index := 0
		length := len(ids)
		for index < length {
			ons = append(ons, table+"."+ids[index]+" = "+withTable+"."+ids[index+1])
			index += 2
		}
	*/

	self.joins = append(self.joins, "JOIN "+withTable+" ON "+strings.Join(ons, " AND "))
}

func (self *SqlWith) Apply(sql string) string {
	if len(self.withs) > 0 {
		sql = "WITH\n" + strings.Join(self.withs, ",\n") + "\n" + sql
		sql = insertSqlBefpre(sql, "GROUP BY", strings.Join(self.joins, "\n"))
	}
	return sql
}

// Utils

func insertSqlBefpre(sql string, place string, insert string) string {
	if insert == "" {
		return sql
	} else if place == "" {
		return sql + "\n" + insert
	} else {
		if p := strings.Index(sql, place); p == -1 {
			return sql + "\n" + insert
		} else {
			return sql[:p] + insert + "\n" + sql[p:]
		}
	}
}

func jsonUnmarshallStringMapEntries(jsonBytes []byte, map_ map[string]string) error {
	if len(jsonBytes) == 0 {
		return nil
	}

	var metadata_ [][]string
	if err := json.Unmarshal(jsonBytes, &metadata_); err == nil {
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

func jsonUnmarshallStringMap(jsonBytes []byte, map_ map[string]string) error {
	if err := json.Unmarshal(jsonBytes, &map_); err == nil {
		return nil
	} else {
		return err
	}
}

func jsonUnmarshallStringArray(jsonBytes []byte, array *[]string) error {
	if len(jsonBytes) > 0 {
		return json.Unmarshal(jsonBytes, array)
	}
	return nil
}

func jsonUnmarshallGvkArray(jsonBytes []byte, gvks *[]util.GVK) error {
	if len(jsonBytes) > 0 {
		var gvks_ [][]string
		if err := json.Unmarshal(jsonBytes, &gvks_); err == nil {
			var gvks__ []util.GVK
			for _, gvk := range gvks_ {
				gvks__ = append(gvks__, util.NewGVK(gvk[0], gvk[1], gvk[2]))
			}
			*gvks = gvks__
		} else {
			return err
		}
	}
	return nil
}
