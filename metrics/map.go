package metrics

type mapElement struct {
	key   []string
	value string
}

type tagMap struct {
	index map[uint64][]mapElement
	hash  func([]string) uint64
}

func newTagMap() *tagMap {
	return &tagMap{
		index: make(map[uint64][]mapElement),
		hash:  hashStringSlice,
	}
}

func (m *tagMap) Add(key []string, value string) {
	hash := m.hash(key)
	i := m.findElem(hash, key)
	if i >= 0 {
		m.index[hash][i].value = value
		return
	}
	m.index[hash] = append(m.index[hash], mapElement{key, value})
}

func (m *tagMap) Get(key []string) (string, bool) {
	hash := m.hash(key)
	i := m.findElem(hash, key)
	if i >= 0 {
		return m.index[hash][i].value, true
	}
	return "", false
}

func (m *tagMap) Delete(key []string) {
	hash := m.hash(key)
	i := m.findElem(hash, key)
	if i >= 0 {
		list := m.index[hash]
		if len(list) > 1 {
			list[i] = list[len(list)-1]
			list[len(list)-1] = mapElement{}
			m.index[hash] = list[:len(list)-1]
		} else {
			delete(m.index, hash)
		}
	}
}

func (m *tagMap) findElem(hash uint64, key []string) int {
	elems := m.index[hash]
	for i := range elems {
		if equalStrings(key, elems[i].key) {
			return i
		}
	}
	return -1
}
