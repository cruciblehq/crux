package manifest

import (
	"errors"
	"os"
	"testing"
)

func TestAsCorrectType(t *testing.T) {
	m := &Manifest{Config: &Widget{Main: "index.js"}}
	w, err := As[*Widget](m)
	if err != nil {
		t.Fatal(err)
	}
	if w.Main != "index.js" {
		t.Fatalf("Main = %q, want %q", w.Main, "index.js")
	}
}

func TestAsWrongType(t *testing.T) {
	m := &Manifest{Config: &Widget{Main: "index.js"}}
	_, err := As[*Runtime](m)
	if !errors.Is(err, ErrConfigTypeMismatch) {
		t.Fatalf("err = %v, want ErrConfigTypeMismatch", err)
	}
}

func TestReadWriteRoundtrip(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	f, err := os.CreateTemp(t.TempDir(), "manifest-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()

	if err := Write(m, path); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Resource.Name != "ns/w" {
		t.Fatalf("Name = %q, want %q", got.Resource.Name, "ns/w")
	}
	w, err := As[*Widget](got)
	if err != nil {
		t.Fatal(err)
	}
	if w.Main != "index.js" {
		t.Fatalf("Main = %q, want %q", w.Main, "index.js")
	}
}

func TestReadWriteAtRoundtrip(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	if err := WriteAt(m, dir); err != nil {
		t.Fatal(err)
	}
	got, err := ReadAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Resource.Name != "ns/w" {
		t.Fatalf("Name = %q, want %q", got.Resource.Name, "ns/w")
	}
}

func TestReadAsRoundtrip(t *testing.T) {
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	f, err := os.CreateTemp(t.TempDir(), "manifest-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	path := f.Name()
	f.Close()

	if err := Write(m, path); err != nil {
		t.Fatal(err)
	}
	w, err := ReadAs[*Widget](path)
	if err != nil {
		t.Fatal(err)
	}
	if w.Main != "index.js" {
		t.Fatalf("Main = %q, want %q", w.Main, "index.js")
	}
}

func TestReadAsAtRoundtrip(t *testing.T) {
	dir := t.TempDir()
	m := &Manifest{
		Resource: Resource{Type: TypeWidget, Name: "ns/w", Version: "1.0.0"},
		Config:   &Widget{Main: "index.js"},
	}
	if err := WriteAt(m, dir); err != nil {
		t.Fatal(err)
	}
	w, err := ReadAsAt[*Widget](dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.Main != "index.js" {
		t.Fatalf("Main = %q, want %q", w.Main, "index.js")
	}
}

func TestWritePlan(t *testing.T) {
	p := &Plan{
		Version:     PlanVersion,
		Services:    map[string]string{"svc": "ns/x"},
		Compute:     map[string]Compute{"c1": {Provider: "local"}},
		Containers:  map[string]Container{"svc": {}},
		Deployments: []Deployment{{Service: "svc", Container: "svc", Compute: "c1"}},
		Gateway:     Gateway{Routes: []Route{{Pattern: "/api", Service: "svc"}}},
	}
	path := t.TempDir() + "/plan.yaml"
	if err := WritePlan(p, path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty plan file")
	}
}

func TestWritePlanAt(t *testing.T) {
	dir := t.TempDir()
	p := &Plan{
		Version:     PlanVersion,
		Services:    map[string]string{"svc": "ns/x"},
		Compute:     map[string]Compute{"c1": {Provider: "local"}},
		Containers:  map[string]Container{"svc": {}},
		Deployments: []Deployment{{Service: "svc", Container: "svc", Compute: "c1"}},
		Gateway:     Gateway{Routes: []Route{{Pattern: "/api", Service: "svc"}}},
	}
	if err := WritePlanAt(p, dir); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(dir + "/plan.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty plan file")
	}
}
