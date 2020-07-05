package utils

import "testing"

func TestCoalesceHttp(t *testing.T) {
	url, err := Coalesce("http://example.com")

	if err != nil {
		t.Errorf("Got err, expected ok. %v", err)
	}

	if url != "http://example.com" {
		t.Errorf("Coalesce(\"http://example.com\") = %s. Expected \"http://example.com\"", url)
	}
}

func TestCoalesceHttps(t *testing.T) {
	url, err := Coalesce("https://example.com")

	if err != nil {
		t.Errorf("Got err, expected ok. %v", err)
	}

	if url != "https://example.com" {
		t.Errorf("Coalesce(\"https://example.com\") = %s. Expected \"https://example.com\"", url)
	}
}

func TestCoalesceNoScheme(t *testing.T) {
	url, err := Coalesce("example.com")

	if err != nil {
		t.Errorf("Got err, expected ok. %v", err)
	}

	if url != "http://example.com" {
		t.Errorf("Coalesce(\"http://example.com\") = %s. Expected \"http://example.com\"", url)
	}
}
