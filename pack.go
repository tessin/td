package td

import (
	"archive/zip"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type PackPattern struct {
	Include []string
	Exclude []string
}

func matchAny(patterns []string, name string) (bool, error) {
	for _, pattern := range patterns {
		if ok, err := filepath.Match(pattern, name); err == nil {
			return ok, nil
		} else {
			return false, err
		}
	}
	return false, nil
}

func (pp *PackPattern) Excludes(name string) (bool, error) {
	if ok, err := matchAny(pp.Exclude, name); err == nil {
		return ok, nil
	} else {
		return false, err
	}
}

func NewPatternParser(filename string, b []byte) *scanner.Scanner {
	fs := token.NewFileSet()
	f := fs.AddFile(filename, fs.Base(), len(b))
	var scanner scanner.Scanner
	scanner.Init(f, b, nil, 0)
	return &scanner
}

func Pack() error {
	// .td-package
	packageFile, err := filepath.Abs(".td-pack")
	if err != nil {
		return err
	}

	// filepath.Walk()

	b, err := ioutil.ReadFile(packageFile)
	if err != nil {
		return err
	}

	var (
		pattern PackPattern
	)

	s := NewPatternParser(packageFile, b)
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		if tok == token.SEMICOLON {
			continue
		}

		if tok == token.NOT {
			pos, tok, lit = s.Scan()
			if tok == token.STRING {
				s, err := strconv.Unquote(lit)
				if err != nil {
					return err
				}
				pattern.Exclude = append(pattern.Exclude, filepath.ToSlash(s))
				continue
			}
		}

		if tok != token.STRING {
			return fmt.Errorf("expected STRING at %v got %v", pos, tok)
		}

		s, err := strconv.Unquote(lit)
		if err != nil {
			return err
		}

		pattern.Include = append(pattern.Include, filepath.ToSlash(s))
	}

	// ====

	log.Println("include", pattern.Include)
	log.Println("exclude", pattern.Exclude)

	// ====

	dir := filepath.Dir(packageFile)

	var glob []string

	for _, include := range pattern.Include {
		matches, err := filepath.Glob(include)
		if err != nil {
			return err
		}
		for _, path := range matches {
			stat, err := os.Stat(path)
			if err != nil {
				return err
			}
			if stat.IsDir() {
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					stat, err := os.Stat(path)
					if err != nil {
						return err
					}
					if stat.IsDir() {
						return nil
					}
					path = filepath.ToSlash(path)
					if exclude, err := pattern.Excludes(path); err == nil {
						if !exclude {
							glob = append(glob, path)
						}
					} else {
						return err
					}
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				path = filepath.ToSlash(path)
				if exclude, err := pattern.Excludes(path); err == nil {
					if !exclude {
						glob = append(glob, path)
					}
				} else {
					return err
				}
			}
		}
	}

	// filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
	// 	f, err := os.Stat(path)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if f.IsDir() {
	// 		return nil
	// 	}
	// 	name, err := filepath.Rel(dir, path)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	name = filepath.ToSlash(name)
	// 	if ok, err := pattern.Matches(name); ok {
	// 		glob = append(glob, name)
	// 	} else {
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}
	// 	return nil
	// })

	log.Println("zippping", len(glob), "files...")

	zipFilename := filepath.Base(dir) + ".zip"

	f, err := os.Create(zipFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	z := zip.NewWriter(f)
	defer z.Close()

	for _, filename := range glob {
		w, err := z.Create(filename)
		if err != nil {
			log.Println("cannot create zip entry", filename)
			return err
		}
		r, err := os.Open(filename)
		if err != nil {
			log.Println("cannot open file", filename)
			return err
		}
		_, err = io.Copy(w, r)
		if err != nil {
			log.Println("cannot copy file into zip archive", filename)
			return err
		}
	}

	return nil
}
