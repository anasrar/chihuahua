package utils

import "math"

func Lerp(v0, v1, t float32) float32 {
	return (1-t)*v0 + t*v1
}

// NOTE: direct port from https://github.com/WoefulWolf/NieR2Blender2NieR/blob/b86ec20312de2c717d65ecaacbd85e7df79e3ab3/utils/ioUtils.py#L87
func PgHalfFloat32FromUint16(num uint16) float32 {
	sign := uint32(num & 0b1000000000000000)
	expo := uint32(num & 0b0111111000000000)
	mant := uint32(num & 0b0000000111111111)

	expo >>= 9

	if expo == 0 && mant == 0 {
		return 0.0
	}

	if expo == 63 {
		if mant == 0 {
			if sign == 1 {
				return float32(math.Inf(-1))
			} else {
				return float32(math.Inf(1))
			}
		} else {
			return float32(math.NaN())
		}
	}

	expo -= 47
	sign <<= 16
	expo += 127
	expo <<= 23
	mant <<= 14

	return math.Float32frombits(sign | expo | mant)
}
