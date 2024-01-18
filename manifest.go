package gocrx

type Manifest struct {
	Name            string `json:"name"`
	ShortName       string `json:"short_name"`
	Description     string `json:"description"`
	Version         string `json:"version"`
	Key             string `json:"key"`
	ManifestVersion int    `json:"manifest_version"`
}
