package audit

import (
	"strings"

	"github.com/C5rogers/G-Synch/internal/audit/core"
)

func mapTables(tables []core.Table) map[string]core.Table {
	m := make(map[string]core.Table, len(tables))
	for _, t := range tables {
		key := strings.ToLower(t.Name)
		m[key] = t
	}
	return m
}

func tableExists(schema *core.Schema, name string) bool {
	for _, t := range schema.Tables {
		if t.Name == name {
			return true
		}
	}
	return false
}

func mapColumns(cols []core.Column) map[string]core.Column {
	m := make(map[string]core.Column, len(cols))
	for _, c := range cols {
		m[strings.ToLower(c.Name)] = c
	}
	return m
}
