package lib

import (
	"encoding/json"
	"errors"
	"fmt"
)

type manifestWrapper struct {
	// DataHash of dataJson
	DataHash string `json:"dataHash"`

	DataJson string `json:"dataJson"`
}

func DeserializePartition(
	dirPath string,
	manifestBytes []byte,
) (*Partition, error) {
	manifest, err := deserializeManifest(manifestBytes)

	if err != nil {
		return nil, err
	}

	p := Partition{
		AbsoluteDirOsPath: dirPath,
		manifest:          manifest,
	}

	return &p, nil
}

func deserializeManifest(jsonBytes []byte) (*manifest, error) {
	wrapper, err := deserializeManifestWrapper(jsonBytes)

	if err != nil {
		fullErr := errors.Join(
			errors.New("while loading serialized manifest"),
			err,
		)

		return nil, fullErr
	}

	manifest, err := wrapper.unwrap()

	if err != nil {
		fullErr := errors.Join(
			errors.New("while parsing .dataJson"),
			err,
		)

		return nil, fullErr
	}

	return manifest, nil
}

func deserializeManifestWrapper(bytes []byte) (*manifestWrapper, error) {
	var wrapper manifestWrapper

	if err := json.Unmarshal(bytes, &wrapper); err != nil {
		return nil, err
	}

	if err := wrapper.validate(); err != nil {
		return nil, err
	}

	if err := wrapper.verifyHash(); err != nil {
		return nil, err
	}

	return &wrapper, nil
}

func (wrapper *manifestWrapper) validate() error {
	if wrapper.DataHash == "" {
		return errors.New(".dataHash must not be empty")
	}

	if wrapper.DataJson == "" {
		return errors.New(".dataJson must not be empty")
	}

	return nil
}

func (wrapper *manifestWrapper) verifyHash() error {
	actualHash := HashString(wrapper.DataJson)

	if actualHash != wrapper.DataHash {
		return fmt.Errorf(
			"manifest hash mismatch. Actual: (%s). Expected (%s)",
			actualHash,
			wrapper.DataHash,
		)
	}

	return nil
}

func (wrapper *manifestWrapper) unwrap() (*manifest, error) {
	var manifest *manifest

	if err := json.Unmarshal([]byte(wrapper.DataJson), &manifest); err != nil {
		return nil, err
	}

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	return manifest, nil
}

func (manifest *manifest) validate() error {
	for path, entry := range manifest.Files {
		err := entry.validate()

		if err == nil {
			continue
		}

		fullErr := errors.Join(
			fmt.Errorf("error in entry %s", path),
			err,
		)

		return fullErr
	}

	return nil
}

func (entry *fileEntry) validate() error {
	if entry.Hash == "" {
		return errors.New(".hash must not be empty")
	}

	if entry.Mtime < 0 {
		return errors.New(".mtime must be >= 0")
	}

	return nil
}

func (partition *Partition) Serialize() ([]byte, error) {
	wrapper, err := partition.manifest.wrap()

	if err != nil {
		return nil, err
	}

	return json.Marshal(wrapper)
}

func (manifest *manifest) wrap() (*manifestWrapper, error) {
	dataJsonBytes, err := json.Marshal(manifest)

	if err != nil {
		return nil, err
	}

	dataJsonString := (string)(dataJsonBytes)

	wrapper := manifestWrapper{
		DataHash: HashString(dataJsonString),
		DataJson: dataJsonString,
	}

	return &wrapper, nil
}
