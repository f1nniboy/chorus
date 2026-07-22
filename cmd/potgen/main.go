package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

type occurrence struct {
	file string
	line int
}

func main() {
	root, err := moduleRoot()
	if err != nil {
		log.Fatal(err)
	}

	fset := token.NewFileSet()
	msgs := map[string][]occurrence{}

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" || path == filepath.Join(root, "cmd", "potgen") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.CallExpr:
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Get" {
					return true
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "locale" {
					return true
				}
				if len(node.Args) == 0 {
					return true
				}
				lit, ok := node.Args[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				addOccurrence(msgs, fset, rel, lit.Pos(), lit.Value)

			case *ast.StructType:
				for _, field := range node.Fields.List {
					if field.Tag == nil {
						continue
					}
					tagVal, err := strconv.Unquote(field.Tag.Value)
					if err != nil {
						continue
					}
					label := reflect.StructTag(tagVal).Get("label")
					if label == "" {
						continue
					}
					addOccurrence(msgs, fset, rel, field.Tag.Pos(), strconv.Quote(label))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	ids := make([]string, 0, len(msgs))
	for id := range msgs {
		ids = append(ids, id)
	}
	slices.Sort(ids)

	var b strings.Builder
	b.WriteString("msgid \"\"\n")
	b.WriteString("msgstr \"\"\n")
	b.WriteString("\"Project-Id-Version: chorus\\n\"\n")
	b.WriteString("\"Content-Type: text/plain; charset=UTF-8\\n\"\n")
	b.WriteString("\"Content-Transfer-Encoding: 8bit\\n\"\n")

	for _, id := range ids {
		refs := make([]string, len(msgs[id]))
		for i, o := range msgs[id] {
			refs[i] = fmt.Sprintf("%s:%d", o.file, o.line)
		}
		fmt.Fprintf(&b, "\n#: %s\n", strings.Join(refs, " "))
		fmt.Fprintf(&b, "msgid \"%s\"\nmsgstr \"\"\n", id)
	}

	if err := os.WriteFile(filepath.Join(root, "data", "po", "default.pot"), []byte(b.String()), 0o644); err != nil {
		log.Fatal(err)
	}
}

func addOccurrence(msgs map[string][]occurrence, fset *token.FileSet, file string, pos token.Pos, quoted string) {
	msgid, err := strconv.Unquote(quoted)
	if err != nil {
		return
	}
	position := fset.Position(pos)
	msgs[msgid] = append(msgs[msgid], occurrence{file: file, line: position.Line})
}

// `go generate` runs from the directory of the file containing the directive,
// not the repo root, so we have to derive the root ourselves
func moduleRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("could not determine source location")
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(file))), nil
}
