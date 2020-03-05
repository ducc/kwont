package utils

func PoundsToMicros(pounds float64) int64 {
	return int64(pounds * 1000000)
}

func MicrosToPounds(micros int64) float64 {
	return float64(micros) / 1000000
}
