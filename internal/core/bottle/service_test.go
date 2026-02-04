package bottle

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDecodeSpec(t *testing.T) {
	svc := &BottleService{}

	spec := BottleSpec{Bottle: BottleMeta{Name: "demo"}, Services: map[string]ServiceSpec{"web": {Image: "nginx"}}}
	data, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	got, err := svc.DecodeSpec(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Bottle.Name != "demo" || len(got.Services) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}

	spec = BottleSpec{Bottle: BottleMeta{Name: ""}, Services: map[string]ServiceSpec{"web": {Image: "nginx"}}}
	data, _ = yaml.Marshal(spec)
	if _, err := svc.DecodeSpec(data); err == nil {
		t.Fatalf("expected error for missing bottle name")
	}

	spec = BottleSpec{Bottle: BottleMeta{Name: "demo"}, Services: map[string]ServiceSpec{}}
	data, _ = yaml.Marshal(spec)
	if _, err := svc.DecodeSpec(data); err == nil {
		t.Fatalf("expected error for missing services")
	}
}

func TestBuildStartOrder(t *testing.T) {
	svc := &BottleService{}

	spec := &BottleSpec{Services: map[string]ServiceSpec{
		"db":  {Image: "postgres"},
		"web": {Image: "nginx", DependsOn: []string{"db"}},
	}}

	order, err := svc.BuildStartOrder(spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBefore(order, "db", "web") {
		t.Fatalf("expected db before web, got %v", order)
	}

	cycle := &BottleSpec{Services: map[string]ServiceSpec{
		"a": {DependsOn: []string{"b"}},
		"b": {DependsOn: []string{"a"}},
	}}
	if _, err := svc.BuildStartOrder(cycle); err == nil {
		t.Fatalf("expected error for dependency cycle")
	}

	unknown := &BottleSpec{Services: map[string]ServiceSpec{
		"a": {DependsOn: []string{"missing"}},
	}}
	if _, err := svc.BuildStartOrder(unknown); err == nil {
		t.Fatalf("expected error for unknown dependency")
	}
}

func TestBuildContainerName(t *testing.T) {
	if got := buildContainerName("b1", "web"); got != "b1-web" {
		t.Fatalf("expected name built, got %q", got)
	}
}

func isBefore(order []string, left, right string) bool {
	leftIdx := -1
	rightIdx := -1
	for i, name := range order {
		switch name {
		case left:
			leftIdx = i
		case right:
			rightIdx = i
		}
	}
	return leftIdx >= 0 && rightIdx >= 0 && leftIdx < rightIdx
}
