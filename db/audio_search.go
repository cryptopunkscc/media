package db

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type AudioSearch struct {
	Text    string
	Include map[string][]string
	Exclude map[string][]string
}

func (db *DB) SearchAudio(search AudioSearch) ([]*AudioRow, error) {
	var rows []*AudioRow
	err := search.apply(db.Model(&AudioRow{})).Find(&rows).Error
	return rows, err
}

func (search AudioSearch) apply(db *gorm.DB) *gorm.DB {
	if search.Text != "" {
		like := "%" + search.Text + "%"
		db = db.Where(
			"LOWER(artist) LIKE ? OR LOWER(title) LIKE ? OR LOWER(album) LIKE ?",
			like, like, like,
		)
	}

	for name, vals := range search.Include {
		db = applyAudioTagCondition(db, false, name, vals)
	}

	for name, vals := range search.Exclude {
		db = applyAudioTagCondition(db, true, name, vals)
	}

	return db
}

func applyAudioTagCondition(db *gorm.DB, negate bool, name string, vals []string) *gorm.DB {
	var clauses []string
	var args []interface{}

	for _, v := range vals {
		if name == "year" {
			c, a, ok := buildYearClause(v)
			if !ok {
				continue
			}
			clauses = append(clauses, c)
			args = append(args, a...)
		} else if name == "format" {
			clauses = append(clauses, "LOWER(format) = ?")
			args = append(args, v)
		} else {
			clauses = append(clauses, "LOWER("+name+") LIKE ?")
			args = append(args, "%"+v+"%")
		}
	}

	if len(clauses) == 0 {
		return db
	}

	expr := "(" + strings.Join(clauses, " OR ") + ")"
	if negate {
		return db.Not(expr, args...)
	}
	return db.Where(expr, args...)
}

func buildYearClause(s string) (clause string, args []interface{}, ok bool) {
	if lo, hi, found := strings.Cut(s, "-"); found {
		loY, err1 := strconv.Atoi(lo)
		hiY, err2 := strconv.Atoi(hi)
		if err1 != nil || err2 != nil {
			return
		}
		if loY > hiY {
			loY, hiY = hiY, loY
		}
		return "year BETWEEN ? AND ?", []interface{}{loY, hiY}, true
	}

	y, err := strconv.Atoi(s)
	if err != nil {
		return
	}

	return "year = ?", []interface{}{y}, true
}
