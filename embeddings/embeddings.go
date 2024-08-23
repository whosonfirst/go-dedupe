package embeddings

func asFloat32(data []float64) []float32 {

	e32 := make([]float32, len(data))

	for idx, v := range data {
		// Really, check for max float32here...
		e32[idx] = float32(v)
	}

	return e32
}

func asFloat64(data []float32) []float64 {

	e64 := make([]float64, len(data))

	for idx, v := range data {
		e64[idx] = float64(v)
	}

	return e64
}
