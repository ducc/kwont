package utils

func PoundsToMicros(pounds int64) int64 {
	return pounds * 1000000
}

func MicrosToPounds(micros int64) float64 {
	return float64(micros) / 1000000
}
