package idmapper2

// IDMap is a bidirectional map to store value <> id translations.
// A value could either be a short id, a name and/or a slug.
type IDMap struct {
	idCache  map[string]string
	valCache map[string]string
}

// NewIDMap creates a new instance.
func NewIDMap() *IDMap {
	return &IDMap{
		idCache:  map[string]string{},
		valCache: map[string]string{},
	}
}

// GetID translates a value to an id.
func (idmap *IDMap) GetID(val string) (string, bool) {
	id, ok := idmap.valCache[val]
	return id, ok
}

// GetName translates an id to a value.
func (idmap *IDMap) GetValue(id string) (string, bool) {
	val, ok := idmap.idCache[id]
	return val, ok
}

// Set writes name and value translation.
func (idmap *IDMap) Set(id string, val string) {
	idmap.idCache[id] = val
	idmap.valCache[val] = id
}
