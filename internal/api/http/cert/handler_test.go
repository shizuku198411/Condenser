package cert

import (
	"net/url"
	"strings"
	"testing"

	"crypto/x509"
)

func TestIsSafeID(t *testing.T) {
	cases := []struct {
		value  string
		expect bool
	}{
		{value: "abc-123", expect: true},
		{value: "", expect: true},
		{value: "ABC", expect: false},
		{value: "a_b", expect: false},
		{value: "a.b", expect: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.value, func(t *testing.T) {
			got := isSafeID(tc.value)
			if got != tc.expect {
				t.Fatalf("expected %v, got %v", tc.expect, got)
			}
		})
	}
}

func TestExtractAndValidateSPIFFE(t *testing.T) {
	validURI := mustParseURL(t, "spiffe://raind/container/abc-123")
	csr := &x509.CertificateRequest{URIs: []*url.URL{validURI}}

	gotURI, role, id, err := extractAndValidateSPIFFE(csr, "raind")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotURI.String() != validURI.String() || role != "container" || id != "abc-123" {
		t.Fatalf("unexpected values: uri=%s role=%s id=%s", gotURI, role, id)
	}

	cases := []struct {
		name string
		csr  *x509.CertificateRequest
		err  string
	}{
		{name: "no uri", csr: &x509.CertificateRequest{URIs: []*url.URL{}}, err: "exactly one"},
		{name: "wrong scheme", csr: &x509.CertificateRequest{URIs: []*url.URL{mustParseURL(t, "http://raind/container/abc")}}, err: "scheme"},
		{name: "wrong host", csr: &x509.CertificateRequest{URIs: []*url.URL{mustParseURL(t, "spiffe://other/container/abc")}}, err: "trust domain"},
		{name: "bad path", csr: &x509.CertificateRequest{URIs: []*url.URL{mustParseURL(t, "spiffe://raind/container")}}, err: "path must"},
		{name: "empty id", csr: &x509.CertificateRequest{URIs: []*url.URL{mustParseURL(t, "spiffe://raind/container/")}}, err: "role/id empty"},
		{name: "invalid chars", csr: &x509.CertificateRequest{URIs: []*url.URL{mustParseURL(t, "spiffe://raind/container/abc_123")}}, err: "invalid characters"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := extractAndValidateSPIFFE(tc.csr, "raind")
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tc.err) {
				t.Fatalf("expected error to contain %q, got %q", tc.err, err.Error())
			}
		})
	}
}

func TestIOReadAllLimit(t *testing.T) {
	data := strings.NewReader("hello")
	got, err := ioReadAllLimit(data, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("expected data, got %q", string(got))
	}

	_, err = ioReadAllLimit(strings.NewReader("too-long"), 3)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "request too large") {
		t.Fatalf("expected request too large error, got %v", err)
	}
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return u
}
