// Package generator implements the goenum code generator: it parses Go
// source files, finds enum declarations annotated with //goenum directives,
// validates them and emits deterministic, gofmt-formatted boilerplate.
package generator

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Kind classifies the underlying basic type of an enum.
type Kind int

const (
	KindInt Kind = iota
	KindUint
	KindString
)

// Value is one enum constant.
type Value struct {
	Ident   string // Go identifier, e.g. StatusActive
	Name    string // canonical name, e.g. "ACTIVE"
	Desc    string // description (may be empty)
	Aliases []string
	Lit     string // Go literal for the value, e.g. "1" or `"active"`
	I64     uint64 // numeric value (integer kinds; sign-preserved bits)
	Str     string // string value (string kind only)
	Pos     token.Position
}

// Enum is a validated enum declaration.
type Enum struct {
	TypeName   string
	Underlying string // as written: int, uint64, string, ...
	Kind       Kind
	Flags      bool
	Values     []Value
	Doc        string // type doc comment, for the generated GoDoc header
	Pkg        string // package name of the source file
	Source     string // path of the source file the enum type was declared in
}

type enumType struct {
	spec       *ast.TypeSpec
	underlying string
	kind       Kind
	flags      bool
	doc        string
	file       string // declaring file path
	wanted     bool   // opted in via //goenum:enum, //goenum:flags or a const directive
}

type constEntry struct {
	et      *enumType
	ident   *ast.Ident
	spec    *ast.ValueSpec
	nameIdx int
	dirs    directives
	expr    ast.Expr
	val     constant.Value
}

// directivePrefix marks generator metadata comments.
const directivePrefix = "//goenum:"

// genFileSuffix is the file-name suffix of generated files; they are skipped
// when scanning directories so regeneration never reads its own output.
const genFileSuffix = "_enum_gen.go"

// ParseFiles parses the given file paths and returns every enum declaration
// whose type is declared in one of them, validated. Files are grouped by
// directory and parsed as a package, so an enum's constants may live in a
// different file of the same package than its type declaration. Errors carry
// file:line positions.
func ParseFiles(paths []string) ([]*Enum, error) {
	byDir := map[string][]string{}
	var dirs []string
	for _, p := range paths {
		d := filepath.Dir(p)
		if _, seen := byDir[d]; !seen {
			dirs = append(dirs, d)
		}
		byDir[d] = append(byDir[d], p)
	}
	sort.Strings(dirs)

	var enums []*Enum
	for _, d := range dirs {
		requested := map[string]bool{}
		for _, p := range byDir[d] {
			requested[p] = true
		}
		found, err := parsePackage(d, requested)
		if err != nil {
			return nil, err
		}
		enums = append(enums, found...)
	}
	sort.SliceStable(enums, func(i, j int) bool { return enums[i].TypeName < enums[j].TypeName })
	return enums, nil
}

// ExpandPaths turns CLI arguments (files, directories, or a recursive
// `dir/...` pattern) into the list of Go source files to scan. Directories
// are scanned non-recursively unless written as `dir/...`; test files and
// previously generated files are skipped.
func ExpandPaths(paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	var out []string
	seen := map[string]bool{}
	add := func(c string) {
		if !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	for _, p := range paths {
		if p == "..." || strings.HasSuffix(p, "/...") {
			root := strings.TrimSuffix(p, "...")
			if root == "" || root == "/" {
				root = "."
			}
			root = strings.TrimSuffix(root, "/")
			err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if e.IsDir() {
					return nil
				}
				if isScanFile(e.Name()) {
					add(path)
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("goenum: %w", err)
			}
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("goenum: %w", err)
		}
		if info.IsDir() {
			entries, err := os.ReadDir(p)
			if err != nil {
				return nil, fmt.Errorf("goenum: %w", err)
			}
			for _, e := range entries {
				if !e.IsDir() && isScanFile(e.Name()) {
					add(filepath.Join(p, e.Name()))
				}
			}
		} else {
			add(p)
		}
	}
	sort.Strings(out)
	return out, nil
}

func isScanFile(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && !strings.HasSuffix(name, genFileSuffix)
}

// parsedFile is one parsed source file with its comment map.
type parsedFile struct {
	path string
	file *ast.File
	cmap ast.CommentMap
}

// parsePackage parses every scannable .go file in dir with one FileSet and
// returns the enums whose type is declared in a requested file.
func parsePackage(dir string, requested map[string]bool) ([]*Enum, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("goenum: %w", err)
	}
	fset := token.NewFileSet()
	var files []parsedFile
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && isScanFile(e.Name()) {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("goenum: parse %s: %w", path, err)
		}
		files = append(files, parsedFile{path: path, file: f, cmap: ast.NewCommentMap(fset, f, f.Comments)})
	}

	var types []*enumType
	byName := map[string]*enumType{}
	for _, pf := range files {
		for _, decl := range pf.file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				underlying, ok := basicUnderlying(ts)
				if !ok {
					continue
				}
				kind := KindInt
				switch {
				case underlying == "string":
					kind = KindString
				case strings.HasPrefix(underlying, "uint"):
					kind = KindUint
				}
				flags := false
				if ts.Doc != nil && hasDirective(ts.Doc, "flags") {
					flags = true
				} else if gd.Doc != nil && len(gd.Specs) == 1 && hasDirective(gd.Doc, "flags") {
					flags = true
				}
				wanted := flags
				if ts.Doc != nil && hasDirective(ts.Doc, "enum") {
					wanted = true
				} else if gd.Doc != nil && len(gd.Specs) == 1 && hasDirective(gd.Doc, "enum") {
					wanted = true
				}
				et := &enumType{spec: ts, underlying: underlying, kind: kind, flags: flags, doc: docText(ts.Doc, gd.Doc), file: pf.path, wanted: wanted}
				types = append(types, et)
				byName[ts.Name.Name] = et
			}
		}
	}

	consts := map[*enumType][]*constEntry{}
	extra := map[string]constant.Value{} // non-enum constants usable in expressions
	for _, pf := range files {
		for _, decl := range pf.file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			var iotaVal int64
			var lastExprs []ast.Expr
			var lastType *ast.Ident
			specIdx := 0
			for _, spec := range gd.Specs {
				vs := spec.(*ast.ValueSpec)
				iotaVal = int64(specIdx)
				specIdx++
				if len(vs.Values) == 0 && len(lastExprs) == 0 {
					return nil, posErr(fset, vs.Pos(), "", "const without value cannot start a const block")
				}
				exprs := lastExprs
				if len(vs.Values) > 0 {
					exprs = vs.Values
					lastExprs = vs.Values
				}
				if vs.Type != nil {
					if tn, ok := vs.Type.(*ast.Ident); ok {
						lastType = tn
					}
				}
				for nameIdx, name := range vs.Names {
					tn := vs.Type
					if tn == nil {
						tn = lastType
					}
					et, isEnum := lookupEnumType(tn, byName)
					dirs, err := directivesFor(fset, pf.cmap, vs, nameIdx)
					if err != nil {
						return nil, err
					}
					var expr ast.Expr
					if len(exprs) == 1 {
						expr = exprs[0]
					} else if nameIdx < len(exprs) {
						expr = exprs[nameIdx]
					} else if len(exprs) > 0 {
						return nil, posErr(fset, vs.Pos(), typeNameOf(et), "value count mismatch in const spec")
					}
					if !isEnum {
						if dirs.hasAny() && requested[pf.path] {
							return nil, posErr(fset, name.Pos(), "", fmt.Sprintf(
								"//goenum directive on %s: unsupported underlying type %s (must be a named int, uint or string type declared in this package)",
								name.Name, exprText(tn)))
						}
						// Keep plain constants available to enum expressions.
						if expr != nil {
							if v, err := evalExpr(fset, expr, iotaVal, mergedConsts(consts, extra)); err == nil {
								extra[name.Name] = v
							}
						}
						continue
					}
					if dirs.has("ignore") {
						continue
					}
					if dirs.declares() {
						et.wanted = true // a const directive opts the type in
					}
					if expr == nil {
						return nil, posErr(fset, name.Pos(), et.spec.Name.Name, "const without value cannot start a const block")
					}
					v, err := evalExpr(fset, expr, iotaVal, mergedConsts(consts, extra))
					if err != nil {
						return nil, err
					}
					consts[et] = append(consts[et], &constEntry{et: et, ident: name, spec: vs, nameIdx: nameIdx, dirs: dirs, expr: expr, val: v})
				}
			}
		}
	}

	var enums []*Enum
	for _, et := range types {
		if !requested[et.file] || !et.wanted {
			continue
		}
		entries := consts[et]
		if len(entries) == 0 {
			if et.flags {
				return nil, posErr(fset, et.spec.Pos(), et.spec.Name.Name, "//goenum:flags type declares no constants")
			}
			if et.wanted {
				return nil, posErr(fset, et.spec.Pos(), et.spec.Name.Name, "opted-in enum declares no constants")
			}
			continue
		}
		e, err := buildEnum(fset, et, entries)
		if err != nil {
			return nil, err
		}
		e.Pkg = pkgNameOf(files, et.file)
		e.Source = et.file
		enums = append(enums, e)
	}
	return enums, nil
}

func typeNameOf(et *enumType) string {
	if et == nil {
		return ""
	}
	return et.spec.Name.Name
}

func pkgNameOf(files []parsedFile, path string) string {
	for _, pf := range files {
		if pf.path == path {
			return pf.file.Name.Name
		}
	}
	return ""
}

// mergedConsts builds the identifier table for expression evaluation: all
// enum constants collected so far plus plain package constants.
func mergedConsts(consts map[*enumType][]*constEntry, extra map[string]constant.Value) map[string]constant.Value {
	m := make(map[string]constant.Value, len(extra))
	for k, v := range extra {
		m[k] = v
	}
	for _, entries := range consts {
		for _, e := range entries {
			if e.val != nil {
				m[e.ident.Name] = e.val
			}
		}
	}
	return m
}

// basicUnderlying reports whether a TypeSpec declares a named enum over a
// supported basic type.
func basicUnderlying(ts *ast.TypeSpec) (string, bool) {
	ident, ok := ts.Type.(*ast.Ident)
	if !ok {
		return "", false
	}
	switch ident.Name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"string":
		return ident.Name, true
	}
	return "", false
}

// lookupEnumType resolves a type annotation to a known enum type.
func lookupEnumType(tn ast.Expr, byName map[string]*enumType) (*enumType, bool) {
	id, ok := tn.(*ast.Ident)
	if !ok || id == nil {
		return nil, false
	}
	et, ok := byName[id.Name]
	return et, ok
}

// buildEnum converts evaluated constants into a validated Enum model.
func buildEnum(fset *token.FileSet, et *enumType, entries []*constEntry) (*Enum, error) {
	e := &Enum{
		TypeName:   et.spec.Name.Name,
		Underlying: et.underlying,
		Kind:       et.kind,
		Flags:      et.flags,
		Doc:        et.doc,
	}
	if et.kind == KindString {
		for _, en := range entries {
			if en.val.Kind() != constant.String {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName,
					fmt.Sprintf("string-backed enum values must be string literals, got %v", constantText(en.val)))
			}
		}
	} else if et.flags {
		if et.kind != KindUint {
			return nil, posErr(fset, et.spec.Pos(), e.TypeName, "//goenum:flags requires an unsigned integer underlying type")
		}
	}

	seenName := map[string]*constEntry{}
	seenFold := map[string]*constEntry{}
	seenAlias := map[string]*constEntry{}
	seenFoldAlias := map[string]*constEntry{}
	seenVal := map[constant.Value]*constEntry{}

	for _, en := range entries {
		name := en.dirs.get("name")
		if name == "" {
			name = defaultName(e.TypeName, en.ident.Name)
		}
		if prev, dup := seenName[name]; dup {
			return nil, posErr(fset, en.ident.Pos(), e.TypeName,
				fmt.Sprintf("duplicate enum name %q: already defined for %s", name, prev.ident.Name))
		}
		fold := strings.ToUpper(name)
		if prev, dup := seenFold[fold]; dup && prev != seenName[name] {
			return nil, posErr(fset, en.ident.Pos(), e.TypeName,
				fmt.Sprintf("ambiguous enum name %q conflicts with %s (case-insensitive)", name, prev.ident.Name))
		}
		seenName[name] = en
		seenFold[fold] = en

		if prev, dup := seenVal[en.val]; dup {
			return nil, posErr(fset, en.ident.Pos(), e.TypeName,
				fmt.Sprintf("duplicate enum value %v: already defined for %s", constantText(en.val), prev.ident.Name))
		}
		seenVal[en.val] = en

		if et.flags {
			if en.val.Kind() != constant.Int {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName, "flag values must be integer literals")
			}
			u, _ := constant.Uint64Val(en.val)
			if u == 0 || u&(u-1) != 0 {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName,
					fmt.Sprintf("flag value %d is not a power of two", u))
			}
		}

		v := Value{Ident: en.ident.Name, Name: name, Desc: en.dirs.get("description"), Pos: fset.Position(en.ident.Pos())}
		switch et.kind {
		case KindString:
			v.Str = constant.StringVal(en.val)
			v.Lit = strconv.Quote(v.Str)
		default:
			if en.val.Kind() != constant.Int {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName, "integer-backed enum values must be integers")
			}
			if strings.HasPrefix(et.underlying, "int") {
				i, _ := constant.Int64Val(en.val)
				v.I64 = uint64(i)
				v.Lit = strconv.FormatInt(i, 10)
			} else {
				u, _ := constant.Uint64Val(en.val)
				v.I64 = u
				v.Lit = strconv.FormatUint(u, 10)
			}
		}
		for _, alias := range en.dirs.all("alias") {
			if alias == "" {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName, "empty alias")
			}
			if alias == name {
				continue // alias equal to own name is redundant, allowed
			}
			foldA := strings.ToUpper(alias)
			if prev, dup := seenName[alias]; dup {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName,
					fmt.Sprintf("alias %q of %s collides with name of %s", alias, en.ident.Name, prev.ident.Name))
			}
			if prev, dup := seenFold[foldA]; dup && prev != seenName[name] {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName,
					fmt.Sprintf("alias %q of %s collides case-insensitively with name of %s", alias, en.ident.Name, prev.ident.Name))
			}
			if prev, dup := seenAlias[alias]; dup {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName,
					fmt.Sprintf("duplicate enum alias %q: already defined for %s", alias, prev.ident.Name))
			}
			if prev, dup := seenFoldAlias[foldA]; dup && prev != seenName[name] {
				return nil, posErr(fset, en.ident.Pos(), e.TypeName,
					fmt.Sprintf("ambiguous enum alias %q conflicts with alias of %s (case-insensitive)", alias, prev.ident.Name))
			}
			seenAlias[alias] = en
			seenFoldAlias[foldA] = en
			v.Aliases = append(v.Aliases, alias)
		}
		e.Values = append(e.Values, v)
	}
	return e, nil
}

// defaultName derives a canonical name from a constant identifier by
// stripping the enum type prefix and uppercasing: StatusPending -> PENDING.
func defaultName(typeName, ident string) string {
	s := strings.TrimPrefix(ident, typeName)
	if s == "" {
		s = ident
	}
	s = strings.TrimFunc(s, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' })
	return strings.ToUpper(s)
}

// constantText renders a constant for error messages.
func constantText(v constant.Value) string {
	if v.Kind() == constant.String {
		return strconv.Quote(constant.StringVal(v))
	}
	return v.ExactString()
}

func docText(docs ...*ast.CommentGroup) string {
	for _, d := range docs {
		if d != nil && d.Text() != "" {
			return strings.TrimRight(d.Text(), "\n")
		}
	}
	return ""
}

func hasDirective(doc *ast.CommentGroup, key string) bool {
	for _, c := range doc.List {
		if strings.HasPrefix(strings.TrimSpace(c.Text), directivePrefix+key) {
			return true
		}
	}
	return false
}

// knownDirectives lists every accepted directive key.
var knownDirectives = map[string]bool{
	"name": true, "description": true, "alias": true, "ignore": true, "flags": true, "enum": true,
}

type directives struct {
	values map[string][]string
}

func (d directives) get(key string) string {
	if v := d.values[key]; len(v) > 0 {
		return v[len(v)-1]
	}
	return ""
}

func (d directives) all(key string) []string { return d.values[key] }
func (d directives) has(key string) bool     { return len(d.values[key]) > 0 }
func (d directives) hasAny() bool            { return len(d.values) > 0 }

// declares reports whether the directives opt the enclosing type into
// generation. `ignore` alone is a filter, not a declaration.
func (d directives) declares() bool {
	for k := range d.values {
		if k != "ignore" {
			return true
		}
	}
	return false
}

// exprText renders a type expression for error messages.
func exprText(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok && id != nil {
		return id.Name
	}
	if expr == nil {
		return "<none>"
	}
	return fmt.Sprintf("%T", expr)
}

// directivesFor extracts //goenum directives attached to a const spec. Doc
// comments above the spec and trailing comments on the same line are both
// accepted. A spec with several names shares its doc comment; trailing
// comments are attributed to the first name.
func directivesFor(fset *token.FileSet, cmap ast.CommentMap, vs *ast.ValueSpec, nameIdx int) (directives, error) {
	d := directives{values: map[string][]string{}}
	if nameIdx != 0 {
		return d, nil
	}
	var groups []*ast.CommentGroup
	if vs.Doc != nil {
		groups = append(groups, vs.Doc)
	}
	for _, g := range cmap[vs] {
		// A comment on the same line as the spec is a trailing comment.
		if fset.Position(g.Pos()).Line == fset.Position(vs.End()).Line {
			groups = append(groups, g)
		}
	}
	for _, g := range groups {
		for _, c := range g.List {
			text := strings.TrimSpace(c.Text)
			if !strings.HasPrefix(text, directivePrefix) {
				continue
			}
			body := strings.TrimPrefix(text, directivePrefix)
			key, val, hasVal := strings.Cut(body, "=")
			key = strings.TrimSpace(key)
			if !knownDirectives[key] {
				return d, posErr(fset, c.Pos(), "", fmt.Sprintf("unknown directive %q (expected one of name, description, alias, ignore, flags)", key))
			}
			if key == "flags" || key == "enum" {
				continue // type-level only; accepted on consts as a no-op marker
			}
			if !hasVal && key != "ignore" {
				return d, posErr(fset, c.Pos(), "", fmt.Sprintf("directive %q requires a value", key))
			}
			d.values[key] = append(d.values[key], strings.TrimSpace(val))
		}
	}
	return d, nil
}

func posErr(fset *token.FileSet, pos token.Pos, typeName, msg string) error {
	p := fset.Position(pos)
	if typeName != "" {
		return fmt.Errorf("%s: %s: %s", p, typeName, msg)
	}
	return fmt.Errorf("%s: %s", p, msg)
}
