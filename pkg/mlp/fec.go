package mlp

// XorParity computes byte-wise XOR parity over a group of payloads.
func XorParity(data [][]byte) []byte {
	maxLen := 0
	for _, d := range data {
		if len(d) > maxLen {
			maxLen = len(d)
		}
	}
	parity := make([]byte, maxLen)
	for _, d := range data {
		for i, b := range d {
			parity[i] ^= b
		}
	}
	return parity
}

// RecoverOne recovers a single missing fragment.
func RecoverOne(missingIndex int, data [][]byte, parity []byte) []byte {
	missing := 0
	for _, d := range data {
		if d == nil {
			missing++
		}
	}
	if missing != 1 {
		return nil
	}
	out := make([]byte, len(parity))
	copy(out, parity)
	for i, d := range data {
		if i == missingIndex || d == nil {
			continue
		}
		for j, b := range d {
			if j < len(out) {
				out[j] ^= b
			}
		}
	}
	return out
}
