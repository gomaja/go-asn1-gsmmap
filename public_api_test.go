package gsmmap

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const asn1ImportPrefix = "github.com/gomaja/go-asn1/"

type apiReference struct {
	importPath string
	name       string
}

func (r apiReference) String() string { return r.importPath + "." + r.name }

type apiSource struct {
	file    *ast.File
	imports map[string]string
}

type apiEnumInfo struct {
	types     map[string]bool
	constants map[string]string // constant name to declared type
}

// TestPublicASN1API checks that callers can name every exposed go-asn1 type
// and enum value through this package, and that go doc associates enum values.
func TestPublicASN1API(t *testing.T) {
	fset := token.NewFileSet()
	packages := make(map[string][]apiSource)
	err := filepath.WalkDir(".", func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if filename != "." {
				switch entry.Name() {
				case "tshark_harness", "temp_wireshark", "testdata", "docs", "plans", "vendor":
					return filepath.SkipDir
				}
				if strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_") {
					return filepath.SkipDir
				}
				if _, statErr := os.Stat(filepath.Join(filename, "go.mod")); statErr == nil {
					return filepath.SkipDir
				} else if !os.IsNotExist(statErr) {
					return statErr
				}
			}
			return nil
		}
		if !strings.HasSuffix(filename, ".go") || strings.HasSuffix(filename, "_test.go") {
			return nil
		}
		file, parseErr := parser.ParseFile(fset, filename, nil, parser.SkipObjectResolution)
		if parseErr != nil {
			return parseErr
		}
		imports := make(map[string]string)
		for _, spec := range file.Imports {
			importPath, unquoteErr := strconv.Unquote(spec.Path.Value)
			if unquoteErr != nil {
				return unquoteErr
			}
			name := path.Base(importPath)
			if spec.Name != nil {
				name = spec.Name.Name
			}
			imports[name] = importPath
		}
		key := filepath.Dir(filename) + ":" + file.Name.Name
		packages[key] = append(packages[key], apiSource{file, imports})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// README: CAMEL phase 4 permits only sms-DELIVER and sms-STATUS-REPORT
	// in MT-SMS TPDU criteria, so sms-SUBMIT-REPORT is intentionally absent.
	exceptions := map[apiReference]string{
		{asn1ImportPrefix + "telecom/ss7/gsm_map", "MTSMSTPDUTypeSmsSUBMITREPORT"}: "CAMEL phase 4 excludes sms-SUBMIT-REPORT from MT-SMS TPDU criteria",
	}
	usedExceptions := make(map[apiReference]bool)
	dependencies := make(map[string]apiEnumInfo)
	var violations []string
	for _, sources := range packages {
		aliases := make(map[apiReference]map[string]bool)
		for _, source := range sources {
			for _, decl := range source.file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}
				for _, spec := range gen.Specs {
					typ := spec.(*ast.TypeSpec)
					if !typ.Name.IsExported() || !typ.Assign.IsValid() {
						continue
					}
					ref, ok := source.reference(typ.Type)
					if !ok {
						continue
					}
					if aliases[ref] == nil {
						aliases[ref] = make(map[string]bool)
					}
					aliases[ref][typ.Name.Name] = true
				}
			}
		}
		checkType := func(node ast.Node, source apiSource) {
			if node == nil {
				return
			}
			ast.Inspect(node, func(n ast.Node) bool {
				expr, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if ref, found := source.reference(expr); found && len(aliases[ref]) == 0 {
					violations = append(violations, fmt.Sprintf("%s: unaliased public type %s", fset.Position(expr.Pos()), ref))
				}
				return true
			})
		}
		wrapped := make(map[apiReference]bool)
		for _, source := range sources {
			for _, decl := range source.file.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					if decl.Name.IsExported() && (decl.Recv == nil || apiExportedReceiver(decl.Recv.List[0].Type)) {
						checkType(decl.Type, source)
					}
				case *ast.GenDecl:
					var constantValues []ast.Expr
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							if spec.Name.IsExported() && !spec.Assign.IsValid() {
								checkType(spec.Type, source)
							}
						case *ast.ValueSpec:
							if len(spec.Values) != 0 {
								constantValues = spec.Values
							}
							for i, name := range spec.Names {
								if !name.IsExported() {
									continue
								}
								checkType(spec.Type, source)
								if decl.Tok != token.CONST || i >= len(constantValues) {
									continue
								}
								ref, ok := source.reference(constantValues[i])
								if !ok {
									continue
								}
								info, loaded := dependencies[ref.importPath]
								if !loaded {
									info = readAPIEnums(t, ref.importPath)
									dependencies[ref.importPath] = info
								}
								enumType := info.constants[ref.name]
								if !info.types[enumType] {
									continue
								}
								wrapped[ref] = true
								typeRef := apiReference{ref.importPath, enumType}
								typeName, explicit := spec.Type.(*ast.Ident)
								if !explicit || !aliases[typeRef][typeName.Name] {
									violations = append(violations, fmt.Sprintf("%s: enum constant %s must explicitly declare an exported alias of %s", fset.Position(name.Pos()), name.Name, typeRef))
								}
							}
						}
					}
				}
			}
		}
		for ref := range aliases {
			info, loaded := dependencies[ref.importPath]
			if !loaded {
				info = readAPIEnums(t, ref.importPath)
				dependencies[ref.importPath] = info
			}
			if !info.types[ref.name] {
				continue
			}
			for name, typ := range info.constants {
				value := apiReference{ref.importPath, name}
				if typ != ref.name || wrapped[value] {
					continue
				}
				if exceptions[value] != "" {
					usedExceptions[value] = true
					continue
				}
				violations = append(violations, fmt.Sprintf("enum alias %s: missing wrapper constant for %s", ref, name))
			}
		}
	}
	for ref := range exceptions {
		if !usedExceptions[ref] {
			violations = append(violations, fmt.Sprintf("exception %s no longer applies", ref))
		}
	}
	// The enum checks rely on go-asn1's generated ENUMERATED comments: fail
	// rather than pass vacuously if the scan stops recognising them.
	if info := dependencies[asn1ImportPrefix+"telecom/ss7/gsm_map"]; !info.types["NetworkResource"] || !info.types["NotReachableReason"] {
		violations = append(violations, fmt.Sprintf("go-asn1 enum scan found %d ENUMERATED types without NetworkResource and NotReachableReason", len(info.types)))
	}
	if len(violations) != 0 {
		sort.Strings(violations)
		t.Fatalf("public go-asn1 API violations (%d):\n%s", len(violations), strings.Join(violations, "\n"))
	}
}

func (source apiSource) reference(expr ast.Expr) (apiReference, bool) {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			break
		}
		expr = paren.X
	}
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return apiReference{}, false
	}
	qualifier, ok := selector.X.(*ast.Ident)
	if !ok {
		return apiReference{}, false
	}
	importPath := source.imports[qualifier.Name]
	return apiReference{importPath, selector.Sel.Name}, strings.HasPrefix(importPath, asn1ImportPrefix)
}

func apiExportedReceiver(expr ast.Expr) bool {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.IsExported()
	case *ast.StarExpr:
		return apiExportedReceiver(expr.X)
	case *ast.IndexExpr:
		return apiExportedReceiver(expr.X)
	case *ast.IndexListExpr:
		return apiExportedReceiver(expr.X)
	default:
		return false
	}
}

// Generated ENUMERATED comments distinguish enums from local INTEGER codes.
// Enums use one-line int64 type declarations and typed constants.
// Scan lines rather than retaining the generated package's large syntax trees.
func readAPIEnums(t *testing.T, importPath string) apiEnumInfo {
	t.Helper()
	pkg, err := build.Default.Import(importPath, ".", build.FindOnly)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(pkg.Dir)
	if err != nil {
		t.Fatal(err)
	}
	typePattern := regexp.MustCompile(`^type\s+(\w+)\s+int64\s*(?://.*)?$`)
	constantPattern := regexp.MustCompile(`^\s*(\w+)\s+(\w+)\s*=`)
	info := apiEnumInfo{make(map[string]bool), make(map[string]string)}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		filename := filepath.Join(pkg.Dir, entry.Name())
		file, openErr := os.Open(filename)
		if openErr != nil {
			t.Fatal(openErr)
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 4096), 1024*1024)
		inConstants := false
		previous := ""
		for scanner.Scan() {
			line := scanner.Text()
			if match := typePattern.FindStringSubmatch(line); match != nil && strings.HasPrefix(previous, "// "+match[1]+" represents the ASN.1 ENUMERATED type ") {
				info.types[match[1]] = true
			}
			previous = line
			trimmed := strings.TrimSpace(line)
			if trimmed == "const (" {
				inConstants = true
				continue
			}
			if inConstants && trimmed == ")" {
				inConstants = false
				continue
			}
			if strings.HasPrefix(trimmed, "const ") {
				line = strings.TrimPrefix(trimmed, "const ")
			} else if !inConstants {
				continue
			}
			if match := constantPattern.FindStringSubmatch(line); match != nil {
				info.constants[match[1]] = match[2]
			}
		}
		scanErr := scanner.Err()
		closeErr := file.Close()
		if scanErr != nil {
			t.Fatalf("%s: %v", filename, scanErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	return info
}
