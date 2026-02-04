package bottle

import (
	"reflect"
	"testing"

	"condenser/internal/core/bottle"
	"condenser/internal/core/container"
	"condenser/internal/store/bsm"
)

func TestMapPolicyChain(t *testing.T) {
	cases := []struct {
		name       string
		policyType string
		expect     string
		wantErr    bool
	}{
		{name: "east-west", policyType: "east-west", expect: "RAIND-EW"},
		{name: "north-south observe", policyType: "north-south_observe", expect: "RAIND-NS-OBS"},
		{name: "north-south enforce", policyType: "north-south_enforce", expect: "RAIND-NS-ENF"},
		{name: "unknown", policyType: "unknown", wantErr: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := mapPolicyChain(tc.policyType)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expect {
				t.Fatalf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestValidatePolicyDestination(t *testing.T) {
	cases := []struct {
		name    string
		chain   string
		dest    string
		wantErr bool
	}{
		{name: "east-west container", chain: "RAIND-EW", dest: "bottle-web"},
		{name: "east-west ip", chain: "RAIND-EW", dest: "10.0.0.1", wantErr: true},
		{name: "east-west cidr", chain: "RAIND-EW", dest: "10.0.0.0/24", wantErr: true},
		{name: "north-south obs ip", chain: "RAIND-NS-OBS", dest: "10.0.0.2"},
		{name: "north-south obs container", chain: "RAIND-NS-OBS", dest: "service-a", wantErr: true},
		{name: "north-south enf ip", chain: "RAIND-NS-ENF", dest: "2001:db8::1"},
		{name: "north-south enf container", chain: "RAIND-NS-ENF", dest: "service-b", wantErr: true},
		{name: "unknown chain ignores", chain: "RAIND-UNKNOWN", dest: "service-c"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validatePolicyDestination(tc.chain, tc.dest)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsIPAddress(t *testing.T) {
	cases := []struct {
		value  string
		expect bool
	}{
		{value: "10.0.0.1", expect: true},
		{value: "10.0.0.0/24", expect: true},
		{value: "2001:db8::1", expect: true},
		{value: "not-an-ip", expect: false},
		{value: "10.0.0.256", expect: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			got := isIPAddress(tc.value)
			if got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestResolvePolicyEndpoint(t *testing.T) {
	services := map[string]bottle.ServiceSpec{
		"web": {Image: "nginx:latest"},
	}
	if got := resolvePolicyEndpoint("b1", services, "web"); got != "b1-web" {
		t.Fatalf("expected service to be resolved, got %q", got)
	}
	if got := resolvePolicyEndpoint("b1", services, "10.0.0.1"); got != "10.0.0.1" {
		t.Fatalf("expected value passthrough, got %q", got)
	}
	if got := resolvePolicyEndpoint("b1", nil, "web"); got != "web" {
		t.Fatalf("expected value passthrough when services nil, got %q", got)
	}
}

func TestNormalizePolicyEndpoint(t *testing.T) {
	if got := normalizePolicyEndpoint("b1", "b1-web"); got != "web" {
		t.Fatalf("expected prefix removed, got %q", got)
	}
	if got := normalizePolicyEndpoint("b1", "other"); got != "other" {
		t.Fatalf("expected value passthrough, got %q", got)
	}
}

func TestBuildBottleContainerName(t *testing.T) {
	if got := buildBottleContainerName("b1", "web"); got != "b1-web" {
		t.Fatalf("expected name built, got %q", got)
	}
}

func TestExtractPolicyIds(t *testing.T) {
	policies := []bsm.PolicyInfo{{Id: "a"}, {Id: ""}, {Id: "b"}}
	got := extractPolicyIds(policies)
	want := []string{"a", "b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := extractPolicyIds(nil); got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestToBottleForwards(t *testing.T) {
	forwards := []container.ForwardInfo{{HostPort: 80, ContainerPort: 8080, Protocol: "tcp"}}
	got := toBottleForwards(forwards)
	want := []BottleForwardInfo{{HostPort: 80, ContainerPort: 8080, Protocol: "tcp"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	if got := toBottleForwards(nil); got != nil {
		t.Fatalf("expected nil for empty input, got %v", got)
	}
}

func TestToStoreAndApiServices(t *testing.T) {
	services := map[string]bottle.ServiceSpec{
		"web": {
			Image:     "nginx:latest",
			Command:   []string{"run"},
			Env:       []string{"A=B"},
			Ports:     []string{"80:80"},
			Mount:     []string{"/data"},
			Network:   "bridge",
			Tty:       true,
			DependsOn: []string{"db"},
		},
	}

	stored := toStoreServices(services)
	wantStored := map[string]bsm.ServiceSpec{
		"web": {
			Image:     "nginx:latest",
			Command:   []string{"run"},
			Env:       []string{"A=B"},
			Ports:     []string{"80:80"},
			Mount:     []string{"/data"},
			Network:   "bridge",
			Tty:       true,
			DependsOn: []string{"db"},
		},
	}
	if !reflect.DeepEqual(stored, wantStored) {
		t.Fatalf("stored services mismatch: expected %v, got %v", wantStored, stored)
	}

	api := toApiServices(stored)
	wantApi := map[string]BottleServiceSpec{
		"web": {
			Image:     "nginx:latest",
			Command:   []string{"run"},
			Env:       []string{"A=B"},
			Ports:     []string{"80:80"},
			Mount:     []string{"/data"},
			Network:   "bridge",
			Tty:       true,
			DependsOn: []string{"db"},
		},
	}
	if !reflect.DeepEqual(api, wantApi) {
		t.Fatalf("api services mismatch: expected %v, got %v", wantApi, api)
	}
}
