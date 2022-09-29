// It is an implementation of the hash/fnv's fnv64a
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function

// Same hash function is used in Prometheus source code
// https://github.com/prometheus/client_golang/blob/main/prometheus/fnv.go

package metrics

// Inline and byte-free variant of hash/fnv's fnv64a.

const (
	offset    uint64 = 14695981039346656037
	prime     uint64 = 1099511628211
	stringSep        = "\n"
)

func hashStringSlice(stringList []string) uint64 {
	hash := uint64(offset)
	for _, str := range stringList {
		hash = hashString(hash, str)
		hash = hashString(hash, stringSep)
	}
	return hash
}

// hashAdd hashes a string on top of an existing fnv64 hash value.
func hashString(hash uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		hash *= prime
		hash ^= uint64(s[i])
	}
	return hash
}
