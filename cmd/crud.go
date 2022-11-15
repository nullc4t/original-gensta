/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"fmt"
	"github.com/nullc4t/gensta/pkg/inspector"
	parser2 "github.com/nullc4t/gensta/pkg/parser"
	"github.com/nullc4t/gensta/pkg/templates"
	"github.com/spf13/cobra"
	"go/ast"
	"go/format"
	astparser "go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"os"
	"path/filepath"
	"strings"
)

// crudCmd represents the crud command
var crudCmd = &cobra.Command{
	Use:     "crud file-with-types output-dir/",
	Aliases: []string{"c", "cr"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args:    cobra.ExactArgs(2),
	Example: "gensta gen crud types.go models/",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("crud called")
		srcFile, err := parser2.NewAstra(args[0])
		if err != nil {
			logger.Fatal(err)
		}

		//fset := token.NewFileSet()
		//file, err := parser.ParseFile(fset, args[0], nil, 0)
		//if err != nil {
		//	logger.Fatal(err)
		//}

		tmpl, err := templates.NewRoot()
		if err != nil {
			logger.Fatal(err)
		}

		tmpl, err = tmpl.ParseFiles("templates/crud_repo.tmpl")
		if err != nil {
			logger.Fatal(err)
		}

		//ast.Print(srcFile.FSet, srcFile.ASTFile)

		for _, decl := range srcFile.ASTFile.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				//logger.Println("st.Fields.List[0].Names==nil", st.Fields.List[0].Names == nil)
				id, ok := st.Fields.List[0].Type.(*ast.Ident)
				if len(st.Fields.List) > 0 && st.Fields.List[0].Names == nil && ok && id.Name == "Model" {
					typeName := ts.Name.Name
					typePackageName := strings.ToLower(typeName)
					logger.Println(typeName, typePackageName)
					tmp := new(bytes.Buffer)
					err = tmpl.ExecuteTemplate(tmp, "crud_repo.tmpl", map[string]any{
						"Package": typePackageName,
						"Type":    fmt.Sprintf("%s.%s", srcFile.Package, typeName),
					})
					if err != nil {
						logger.Fatal(err)
					}

					fmt.Println(string(tmp.Bytes()))

					fset := token.NewFileSet()
					file, err := astparser.ParseFile(fset, "", tmp.Bytes(), 0)
					if err != nil {
						logger.Fatal(err)
					}

					ok := astutil.AddImport(fset, file, srcFile.ImportPath())
					if !ok {
						logger.Fatal("not ok")
					}

					for t, _ := range inspector.GetImportedTypes(srcFile.Astra) {
						p := inspector.ExtractPackageFromType(t)
						if importPath := inspector.GetImportPathForPackage(p, srcFile.Astra); importPath != "" {
							astutil.AddImport(fset, file, importPath)
						}
					}

					ast.SortImports(fset, file)

					tmp = new(bytes.Buffer)
					err = printer.Fprint(tmp, fset, file)
					if err != nil {
						logger.Fatal(err)
					}

					formatted, err := format.Source(tmp.Bytes())
					if err != nil {
						logger.Fatal(err)
					}

					writePath := filepath.Join(args[1], typePackageName, "repo.gensta.go")
					f, err := os.OpenFile(writePath, os.O_WRONLY|os.O_CREATE, 0644)
					if os.IsNotExist(err) {
						err = os.MkdirAll(filepath.Dir(writePath), 0755)
						if err != nil {
							logger.Fatal(err)
						}

						f, err = os.Create(writePath)
						if err != nil {
							logger.Fatal(err)
						}

					}
					if err != nil {
						logger.Fatal(err)
					}
					defer f.Close()

					_, err = f.Write(formatted)
					if err != nil {
						logger.Fatal(err)
					}

					fmt.Println("Done")
				}
			}
			//logger.Println(i, decl)
			//ast.Print(fset, decl)
		}
	},
}

func init() {
	genCmd.AddCommand(crudCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// crudCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// crudCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
