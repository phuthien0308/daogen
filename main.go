package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/phuthien0308/daogen/internal"

	"github.com/spf13/cobra"
)

// daogen -t users -s User -o
var (
	tableName  string
	structName string
	output     string
)

var rootCmd = &cobra.Command{
	Use:   "daogen ",
	Short: "a simple dao generator",
	Long:  `a simple dao generator which supports to generate dao files which work with RDS databases`,
	Args:  cobra.ArbitraryArgs,
	Run:   excution,
}

func init() {
	rootCmd.Flags().StringVarP(&tableName, "table", "t", "", "table name")
	rootCmd.Flags().StringVarP(&structName, "struct", "s", "", "struct name")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "output")
}

func excution(cmd *cobra.Command, args []string) {

	tbl := internal.NewTableTag(structName, tableName)
	query, fields := tbl.BuildAddQuery()

	outputPkgName, importSet, includePkgModel := internal.New(structName, output)

	generate(TemplateData{
		OutputPackage: string(outputPkgName),
		Imports:       importSet,

		IsIncludedPkgModelName: bool(includePkgModel),
		Type:                   structName,
		TypeLower:              strings.ToLower(structName),
		InsertFields:           fields,
		InsertQuery:            query,
	})

}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

type TemplateData struct {
	OutputPackage          string
	Imports                internal.ImportSet
	IsIncludedPkgModelName bool
	Type                   string
	TypeLower              string
	InsertFields           []string
	InsertQuery            string
	UpdateQuery            string
	DeleteQuery            string
}

func ensureDir(dir string) (string, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	// Create the directory (and parents) if it doesn't exist
	err = os.MkdirAll(absPath, 0755)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func generate(data TemplateData) {
	tmpl, err := template.ParseFiles("internal/dao.go.tmpl")
	if err != nil { /* handle error */
	}
	// Create the output file (e.g., "user_dao.go")
	dir, err := ensureDir(output)
	if err != nil {
		panic(err)
	}

	if data.OutputPackage == "" {
		data.OutputPackage = filepath.Base(dir)
	}

	var buffer bytes.Buffer
	// Execute the template with the data and write to the file
	err = tmpl.Execute(&buffer, data)

	if err != nil {
	}
	formatted, err := format.Source(buffer.Bytes())

	if err != nil {
		panic(err)
	}
	outFile, err := os.Create(dir + "/" + data.TypeLower + "_dao.go")
	if err != nil {
		panic(err)
	}

	defer outFile.Close()
	_, err = outFile.Write(formatted)
	if err != nil {
		panic(err)
	}
}
