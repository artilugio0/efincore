package efincore

import (
	"net/http"
	"testing"
)

func TestNilCriteriaClone(t *testing.T) {
	var c *criteria
	cloned := c.clone()

	if cloned == nil {
		t.Errorf("did not expect a nil cloned criteria")
	}
}

func TestNilCriteria_ShouldInterceptDomain_ReturnsTrue(t *testing.T) {
	var c *criteria

	got := c.shouldInterceptDomain("www.example.com")
	expected := true

	if got != expected {
		t.Errorf("expected '%t', got '%t'", got, expected)
	}
}

func TestCriteria_ShouldInterceptDomain_ReturnsTrueByDefault(t *testing.T) {
	c := newCriteria()

	got := c.shouldInterceptDomain("www.example.com")
	expected := true

	if got != expected {
		t.Errorf("expected '%t', got '%t'", got, expected)
	}
}

func TestCriteria_ShouldInterceptRequest_ReturnsTrueByDefault(t *testing.T) {
	c := newCriteria()

	r, err := http.NewRequest("GET", "https://www.example.com", nil)
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}

	got := c.shouldInterceptRequest(r)
	expected := true

	if got != expected {
		t.Errorf("expected '%t', got '%t'", got, expected)
	}
}
