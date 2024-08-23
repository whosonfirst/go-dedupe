package embeddings

func asFloat32(data []float64) []float32 {

	e32 := make([]float32, len(data))

	for idx, v := range data {
		e32[idx] = float32(v)
	}

	return e32
}
