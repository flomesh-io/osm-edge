package utils

import "hash/fnv"

// HashFromString calculates an FNV-1 hash from a given string,
// returns it as a uint64 and error, if any
func HashFromString(s string) (uint32, error) {
	h := fnv.New32()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0, err
	}

	return h.Sum32() << 1 >> 1, nil
}
