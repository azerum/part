package main

import (
	"strings"
	"testing"
)

func Test_ParseManifestEnvelopeJson_returns_error_when_passed_string_is_not_a_valid_json(t *testing.T) {
	_, err := ParseManifestEnvelopeJson("{ this is not valid json")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_ParseManifestEnvelopeJson_returns_error_when_data_json_is_missing(t *testing.T) {
	_, err := ParseManifestEnvelopeJson(`{"sha1Hex":"abc"}`)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), ".dataJson") {
		t.Fatalf("expected error mentioning .dataJson, got: %v", err)
	}
}

func Test_ParseManifestEnvelopeJson_returns_error_when_checksum_does_not_match(t *testing.T) {
	envelopeJson := `{"sha1Hex":"0","dataJson":"{}"}`
	expectedHash := sha1HexOfString("{}")

	_, err := ParseManifestEnvelopeJson(envelopeJson)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "0") {
		t.Fatal("error message should contain actual hash")
	}

	if !strings.Contains(err.Error(), expectedHash) {
		t.Fatal("error message should contain expected hash")
	}
}

func Test_LoadManifest_returns_error_when_data_json_is_not_a_valid_json(t *testing.T) {
	envelope := ManifestEnvelope{
		Sha1Hex:  "ignored",
		DataJson: "{ not a json }",
	}

	_, err := envelope.LoadManifest()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_LoadManifest_returns_error_when_any_entry_is_malformed(t *testing.T) {
	// Here, file entry is missing .sha1Hex
	envelope := ManifestEnvelope{
		Sha1Hex: "ignored",

		DataJson: `
			{
				"files": {
					"path/to/file.txt": {
						"mtime": 0
					}
				}
			}
		`,
	}

	_, err := envelope.LoadManifest()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()

	msgContainsEverythingWeNeed :=
		strings.Contains(msg, ".sha1Hex") &&
			strings.Contains(msg, "path/to/file.txt")

	if !msgContainsEverythingWeNeed {
		t.Fatalf("expected error message mentioning all context of the error (see test code)")
	}
}
