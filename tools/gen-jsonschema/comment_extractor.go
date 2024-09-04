// Copyright (C) 2014 Alec Thomas

// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
// of the Software, and to permit persons to whom the Software is furnished to do
// so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Copied from
// https://github.com/invopop/jsonschema/blob/v0.12.0/comment_extractor.go and
// adapted to extract the entire comment from Types and Fields, rather than only the first line.

package main

import (
	"fmt"
	"io/fs"
	gopath "path"
	"path/filepath"
	"strings"

	"go/ast"
	"go/parser"
	"go/token"
)

// ExtractGoComments will read all the go files contained in the provided path,
// including sub-directories, in order to generate a dictionary of comments
// associated with Types and Fields. The results will be added to the `commentsMap`
// provided in the parameters and expected to be used for Schema "description" fields.
//
// The `go/parser` library is used to extract all the comments and unfortunately doesn't
// have a built-in way to determine the fully qualified name of a package. The `base` paremeter,
// the URL used to import that package, is thus required to be able to match reflected types.
//
// When parsing type comments, we use the `go/doc`'s Synopsis method to extract the first phrase
// only. Field comments, which tend to be much shorter, will include everything.
func ExtractGoComments(base, path string, commentMap map[string]string) error {
	fset := token.NewFileSet()
	dict := make(map[string][]*ast.Package)
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			d, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			for _, v := range d {
				// paths may have multiple packages, like for tests
				k := gopath.Join(base, path)
				dict[k] = append(dict[k], v)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for pkg, p := range dict {
		for _, f := range p {
			gtxt := ""
			typ := ""
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.TypeSpec:
					typ = x.Name.String()
					if !ast.IsExported(typ) {
						typ = ""
					} else {
						txt := x.Doc.Text()
						if txt == "" && gtxt != "" {
							txt = gtxt
							gtxt = ""
						}
						// This is the modified line.
						// Don't use the Synopsis which only contains the first
						// line/sentence, use the entire comment.
						// txt = doc.Synopsis(txt)
						commentMap[fmt.Sprintf("%s.%s", pkg, typ)] = strings.TrimSpace(strings.ReplaceAll(txt, "\n", " "))
					}
				case *ast.Field:
					txt := x.Doc.Text()
					if txt == "" {
						txt = x.Comment.Text()
					}
					if typ != "" && txt != "" {
						for _, n := range x.Names {
							if ast.IsExported(n.String()) {
								k := fmt.Sprintf("%s.%s.%s", pkg, typ, n)
								commentMap[k] = strings.TrimSpace(txt)
							}
						}
					}
				case *ast.GenDecl:
					// remember for the next type
					gtxt = x.Doc.Text()
				}
				return true
			})
		}
	}

	return nil
}
