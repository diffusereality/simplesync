package main

type Manifest struct {
	Path string
	Body map[string]any
}

func (m Manifest) ExtractMetadata() (kind, name string) {
	if k, ok := m.Body["kind"].(string); ok {
		kind = k
	}
	if n, ok := m.Body["name"].(string); ok {
		name = n
	}

	return
}

func (m Manifest) Kind() string {
	k, _ := m.ExtractMetadata()
	return k
}
