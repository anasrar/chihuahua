package utils

func Lerp(v0, v1, t float32) float32 {
	return (1-t)*v0 + t*v1
}
