package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
)

type ManifestEnvelope struct {
	Sha1Hex  string `json:"sha1Hex"`
	DataJson string `json:"dataJson"`
}

type Manifest struct {
	Files map[string]FileEntry `json:"files"`
}

type FileEntry struct {
	Sha1Hex string `json:"sha1Hex"`
	Mtime   int64  `json:"mtime"`
}

func ParseManifestEnvelopeJson(jsonString string) (*ManifestEnvelope, error) {
	var envelope ManifestEnvelope

	if err := json.Unmarshal([]byte(jsonString), &envelope); err != nil {
		return nil, err
	}

	if err := envelope.validateFields(); err != nil {
		return nil, err
	}

	if err := envelope.checksum(); err != nil {
		return nil, err
	}

	return &envelope, nil
}

func (envelope *ManifestEnvelope) LoadManifest() (*Manifest, error) {
	var manifest Manifest

	if err := json.Unmarshal([]byte(envelope.DataJson), &manifest); err != nil {
		return nil, err
	}

	if err := manifest.validateFields(); err != nil {
		return nil, err
	}

	return &manifest, nil

}

func (envelope *ManifestEnvelope) validateFields() error {
	if envelope.Sha1Hex == "" {
		return errors.New(".sha1Hex must not be empty")
	}

	if envelope.DataJson == "" {
		return errors.New(".dataJson must not be empty")
	}

	return nil
}

func (manifest *Manifest) validateFields() error {
	for path, entry := range manifest.Files {
		if path == "" {
			return errors.New("one of the .files entries has empty path")
		}

		err := entry.validateFields()

		if err != nil {
			return errors.Join(
				fmt.Errorf("while parsing entry for path: %s", path),
				err,
			)
		}
	}

	return nil
}

func (entry *FileEntry) validateFields() error {
	if entry.Sha1Hex == "" {
		return errors.New(".sha1Hex must not be empty")
	}

	if entry.Mtime < 0 {
		return errors.New(".mtime must be non-negative")
	}

	return nil
}

func (envelope *ManifestEnvelope) checksum() error {
	actualSha1 := sha1HexOfString(envelope.DataJson)

	if actualSha1 != envelope.Sha1Hex {
		return fmt.Errorf(
			"actual SHA1 of .dataJson (1) does not match expected (2)\n"+
				"1: %s, 2: %s",

			actualSha1,
			envelope.Sha1Hex,
		)
	}

	return nil
}

func sha1HexOfString(s string) string {
	sum := sha1.Sum([]byte(s))
	return fmt.Sprintf("%x", sum)
}

func (manifest *Manifest) ToEnvelope() ManifestEnvelope {
	jsonBytes, err := json.Marshal(manifest)

	if err != nil {
		panic(err)
	}

	jsonString := (string)(jsonBytes)

	return ManifestEnvelope{
		Sha1Hex:  sha1HexOfString(jsonString),
		DataJson: jsonString,
	}
}
