package template

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// Implements fs.DirEntry for directory tests
type mockDirEntry struct {
	name  string
	isDir bool
}

type mockTemplateInfo struct {
	Check string `yaml:"check"`
}

func (e *mockDirEntry) Name() string { return e.name }
func (e *mockDirEntry) IsDir() bool  { return e.isDir }
func (e *mockDirEntry) Type() fs.FileMode {
	if e.isDir {
		return fs.ModeDir
	}
	return 0
}
func (e *mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

// executeTemplate tests

func TestExecuteTemplate_Success(t *testing.T) {
	tmpl := "Hello, {{.Name}}!"
	data := struct{ Name string }{Name: "Crucible"}
	out, err := executeTemplate("(templateFile)", "test.tmpl", []byte(tmpl), data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "Hello, Crucible!" {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestExecuteTemplate_ParseError(t *testing.T) {
	_, err := executeTemplate("(templateFile)", "bad.tmpl", []byte("{{.Name"), struct{ Name string }{"Crucible"})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
}

func TestExecuteTemplate_ExecuteError(t *testing.T) {
	_, err := executeTemplate("(templateFile)", "bad.tmpl", []byte("{{.Name}}"), struct{}{})
	if err == nil {
		t.Fatal("expected error for missing field")
	}
}

// instantiateTemplateEntry tests

func TestInstantiateTemplateEntry_Directory(t *testing.T) {
	tmp := t.TempDir()
	entry := &mockDirEntry{name: "subdir", isDir: true}
	options := &TemplateOptions{DirMode: 0755, Location: tmp}
	err := instantiateTemplateEntry(nil, "", entry, filepath.Join(tmp, "subdir"), options)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	info, err := os.Stat(filepath.Join(tmp, "subdir"))
	if err != nil || !info.IsDir() {
		t.Errorf("directory not created")
	}
}

func TestInstantiateTemplateEntry_File(t *testing.T) {
	fileName := "file.txt"
	tmp := t.TempDir()
	fsMap := fstest.MapFS{
		fileName: &fstest.MapFile{Data: []byte("content")},
	}
	entry := &mockDirEntry{name: fileName, isDir: false}
	options := &TemplateOptions{FileMode: 0644, Location: tmp}
	outPath := filepath.Join(tmp, fileName)
	err := instantiateTemplateEntry(fsMap, fileName, entry, outPath, options)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil || string(data) != "content" {
		t.Errorf("file not written correctly")
	}
}

func TestInstantiateTemplateEntry_Template(t *testing.T) {
	fileName := "file.tmpl"
	tmp := t.TempDir()
	fsMap := fstest.MapFS{
		fileName: &fstest.MapFile{Data: []byte("Hello, {{.Name}}!")},
	}
	entry := &mockDirEntry{name: fileName, isDir: false}
	options := &TemplateOptions{FileMode: 0644, Location: tmp, Data: struct{ Name string }{"Test"}}
	outPath := filepath.Join(tmp, "file")
	err := instantiateTemplateEntry(fsMap, fileName, entry, outPath, options)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil || !strings.Contains(string(data), "Hello, Test") {
		t.Errorf("template not processed correctly")
	}
}

// Create tests

func TestCreate_Success(t *testing.T) {
	tmp := t.TempDir()
	outDir := filepath.Join(tmp, "output")
	fsMap := fstest.MapFS{
		"tmpl/file.txt": &fstest.MapFile{Data: []byte("data")},
	}
	options := &TemplateOptions{
		Resource: "resource",
		Template: "tmpl",
		Location: outDir,
		FileMode: 0644,
		DirMode:  0755,
	}

	err := Create[mockTemplateInfo](fsMap, options)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "resource/file.txt"))
	if err != nil || string(data) != "data" {
		t.Errorf("file not created correctly")
	}
}

func TestCreate_SkipsMetafile(t *testing.T) {
	tmp := t.TempDir()
	fsMap := fstest.MapFS{
		"tmpl/file.txt":  &fstest.MapFile{Data: []byte("content")},
		"tmpl/.meta.yml": &fstest.MapFile{Data: []byte("metadata")},
	}
	options := &TemplateOptions{
		Resource: "output",
		Template: "tmpl",
		Location: tmp,
		Metafile: ".meta.yml",
		FileMode: 0644,
		DirMode:  0755,
	}
	err := Create[mockTemplateInfo](fsMap, options)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "output/.meta.yml")); !os.IsNotExist(err) {
		t.Error("metafile should not be copied to output")
	}
	if _, err := os.Stat(filepath.Join(tmp, "output/file.txt")); err != nil {
		t.Error("regular file should be copied to output")
	}
}

func TestCreate_PathExists(t *testing.T) {
	tmp := t.TempDir()
	options := &TemplateOptions{Location: tmp}
	err := Create[mockTemplateInfo](nil, options)
	if err == nil {
		t.Fatal("expected error when output path exists")
	}
}

func TestCreate_PathExistsReturnsOutputError(t *testing.T) {
	tmp := t.TempDir()
	fsMap := fstest.MapFS{
		"tmpl/file.txt": &fstest.MapFile{Data: []byte("data")},
	}
	options := &TemplateOptions{
		Resource: filepath.Base(tmp),
		Template: "tmpl",
		Location: filepath.Dir(tmp),
	}
	err := Create[mockTemplateInfo](fsMap, options)
	if err == nil {
		t.Fatal("expected error when output path exists")
	}
	if !errors.Is(err, ErrOutput) {
		t.Errorf("expected ErrOutput, got %v", err)
	}
	if !errors.Is(err, fs.ErrExist) {
		t.Errorf("expected fs.ErrExist, got %v", err)
	}
}

func TestCreate_MissingTemplateReturnsTemplateError(t *testing.T) {
	tmp := t.TempDir()
	fsMap := fstest.MapFS{}
	options := &TemplateOptions{
		Resource: "output",
		Template: "nonexistent",
		Location: tmp,
	}
	err := Create[mockTemplateInfo](fsMap, options)
	if err == nil {
		t.Fatal("expected error for missing template")
	}
	if !errors.Is(err, ErrTemplate) {
		t.Errorf("expected ErrTemplate, got %v", err)
	}
}

// List tests

func TestList_Success(t *testing.T) {
	fsMap := fstest.MapFS{
		"tmpl/meta": &fstest.MapFile{Data: []byte("metadata")},
	}
	decoder := func(data []byte, info *mockTemplateInfo) error {
		info.Check = string(data)
		return nil
	}
	templates, err := List(fsMap, "meta", decoder)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(templates) != 1 || templates[0].Name != "tmpl" {
		t.Errorf("unexpected template list result")
	}
}

func TestList_DecodeError(t *testing.T) {
	fsMap := fstest.MapFS{
		"tmpl/meta": &fstest.MapFile{Data: []byte("metadata")},
	}
	decoder := func(data []byte, info *mockTemplateInfo) error {
		return fmt.Errorf("decode error")
	}
	_, err := List(fsMap, "meta", decoder)
	if err == nil {
		t.Fatal("expected error from decoder")
	}
}

// Render tests

func TestRender_Success(t *testing.T) {
	fsMap := fstest.MapFS{
		"test.tmpl": &fstest.MapFile{Data: []byte("Hello, {{.Name}}!")},
	}
	var buf strings.Builder
	err := Render(fsMap, "test.tmpl", &buf, struct{ Name string }{"World"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if buf.String() != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %s", buf.String())
	}
}

func TestRender_ParseError(t *testing.T) {
	fsMap := fstest.MapFS{
		"bad.tmpl": &fstest.MapFile{Data: []byte("{{.Name")},
	}
	var buf strings.Builder
	err := Render(fsMap, "bad.tmpl", &buf, struct{}{})
	if err == nil {
		t.Fatal("expected error for invalid template")
	}
	if !errors.Is(err, ErrTemplate) {
		t.Errorf("expected ErrTemplate, got %v", err)
	}
}

// Template function tests

func TestTemplateFunc_Default_EmptyString(t *testing.T) {
	tmpl := `{{default "fallback" .Value}}`
	out, err := executeTemplate("test", "test.tmpl", []byte(tmpl), struct{ Value string }{""})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "fallback" {
		t.Errorf("expected 'fallback', got %s", out)
	}
}

func TestTemplateFunc_Default_WithValue(t *testing.T) {
	tmpl := `{{default "fallback" .Value}}`
	out, err := executeTemplate("test", "test.tmpl", []byte(tmpl), struct{ Value string }{"actual"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "actual" {
		t.Errorf("expected 'actual', got %s", out)
	}
}

func TestTemplateFunc_Default_Nil(t *testing.T) {
	tmpl := `{{default "fallback" .Value}}`
	out, err := executeTemplate("test", "test.tmpl", []byte(tmpl), struct{ Value interface{} }{nil})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "fallback" {
		t.Errorf("expected 'fallback', got %s", out)
	}
}

func TestTemplateFunc_Default_EmptySlice(t *testing.T) {
	tmpl := `{{default "fallback" .Value}}`
	out, err := executeTemplate("test", "test.tmpl", []byte(tmpl), struct{ Value []string }{[]string{}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "fallback" {
		t.Errorf("expected 'fallback', got %s", out)
	}
}

func TestTemplateFunc_JSON(t *testing.T) {
	tmpl := `{{json .Data}}`
	out, err := executeTemplate("test", "test.tmpl", []byte(tmpl), struct{ Data []string }{[]string{"a", "b"}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != `["a","b"]` {
		t.Errorf("expected '[\"a\",\"b\"]', got %s", out)
	}
}

func TestTemplateFunc_Slice(t *testing.T) {
	tmpl := `{{range slice 1 2 3}}{{.}}{{end}}`
	out, err := executeTemplate("test", "test.tmpl", []byte(tmpl), struct{}{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(out) != "123" {
		t.Errorf("expected '123', got %s", out)
	}
}
