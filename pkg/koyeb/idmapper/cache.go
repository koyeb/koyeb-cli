package idmapper

// Cache is a bidirectional map to store name <> id translations.
type Cache struct {
	idCache   map[string]string
	nameCache map[string]string
}

// NewCache creates a new instance.
func NewCache() *Cache {
	return &Cache{
		idCache:   map[string]string{},
		nameCache: map[string]string{},
	}
}

// GetID translates a name to an id.
func (cache *Cache) GetID(name string) (string, bool) {
	id, ok := cache.nameCache[name]
	return id, ok
}

// GetName translates an id to a name.
func (cache *Cache) GetName(id string) (string, bool) {
	name, ok := cache.idCache[id]
	return name, ok
}

// Set writes name and id translation.
func (cache *Cache) Set(id string, name string) {
	cache.idCache[id] = name
	cache.nameCache[name] = id
}
