package internal

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

const (
	identityTag = "sql-identity"
	updateTag   = "sql-update"
	insertTag   = "sql-insert"
	colTag      = "sql-col"
)

type FieldTag struct {
	FieldName  string
	ColumnName string
	IsIdentity bool
	IsInsert   bool
	IsUpdate   bool
}

type TableTag struct {
	TableName string
	Tags      []FieldTag
}

func NewTableTag(structName, tableName string) TableTag {

	currentDir, _ := os.Getwd()
	tbl := TableTag{}
	filepath.WalkDir(currentDir, func(path string, d fs.DirEntry, err error) error {

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".go" {

			fst := token.NewFileSet()
			f, err := parser.ParseFile(fst, path, nil, parser.AllErrors)
			if err != nil {
				panic(err)
			}
			ast.Inspect(f, func(n ast.Node) bool {
				if inspectType, ok := n.(*ast.TypeSpec); ok {
					structType, isOkay := inspectType.Type.(*ast.StructType)
					if !isOkay {
						return true
					}
					if inspectType.Name.Name == structName {
						tbl.TableName = tableName
						for _, field := range structType.Fields.List {
							newTag := FieldTag{}
							if len(field.Names) > 0 {
								newTag.FieldName = field.Names[0].Name
								if field.Tag != nil {
									tagString, _ := strconv.Unquote(field.Tag.Value)
									structTag := reflect.StructTag(tagString)
									colname := structTag.Get(colTag)
									newTag.ColumnName = colname
									update := structTag.Get(updateTag)
									newTag.IsIdentity = false
									newTag.IsUpdate = true
									newTag.IsInsert = true
									if update != "" {
										val, err := strconv.ParseBool(update)
										if err != nil {
											return false
										}
										newTag.IsUpdate = val
									}
									insert := structTag.Get(insertTag)
									if insert != "" {
										val, err := strconv.ParseBool(insert)
										if err != nil {
											return false
										}
										newTag.IsInsert = val
									}
									identity := structTag.Get(identityTag)
									if identity != "" {
										val, err := strconv.ParseBool(identity)
										if err != nil {
											return false
										}
										newTag.IsIdentity = val
									}
									tbl.Tags = append(tbl.Tags, newTag)
								}

							}

						}
					}
				}
				return true
			})
		}

		return nil
	})
	return tbl
}

func (tbl TableTag) buildWhere() string {

	var where string
	for _, f := range tbl.Tags {
		if f.IsIdentity {
			if where != "" {
				where = where + " and "
			}
			where = where + fmt.Sprintf("%s = ?", f.ColumnName)
		}
	}
	return where
}

func (tbl TableTag) buildInsertColumns() (string, string, []string) {
	var cols []string
	var questionValues []string
	var properties []string

	for _, tag := range tbl.Tags {
		if !tag.IsIdentity {
			cols = append(cols, tag.ColumnName)
			questionValues = append(questionValues, "?")
			properties = append(properties, tag.FieldName)
		}
	}
	return fmt.Sprintf("(%v)", strings.Join(cols, ",")),
		fmt.Sprintf("(%v)", strings.Join(questionValues, ",")),
		properties
}

func (tbl TableTag) buildUpdateColumns() string {
	var cols []string

	for _, tag := range tbl.Tags {
		if !tag.IsIdentity {
			cols = append(cols, fmt.Sprintf("%v = ?", tag.ColumnName))
		}
	}
	return fmt.Sprintf("%v", strings.Join(cols, ","))
}

func (tbl *TableTag) BuildAddQuery() (string, []string) {
	inserttTmpl := `INSERT INTO %v %v VALUES %v`
	cols, vals, properties := tbl.buildInsertColumns()
	return fmt.Sprintf(inserttTmpl, tbl.TableName, cols, vals), properties
}

func (tbl *TableTag) BuildUpdateQuery() string {
	updateTmpl := "UPDATE %v SET %v WHERE %v"
	return fmt.Sprintf(updateTmpl, tbl.TableName, tbl.buildUpdateColumns(), tbl.buildWhere())
}

func (tbl *TableTag) BuildDeleteQuery() string {
	deleteTmpl := "DELETE FROM %v WHERE %v"
	return fmt.Sprintf(deleteTmpl, tbl.TableName, tbl.buildWhere())
}
