package utils

import rl "github.com/gen2brain/raylib-go/raylib"

// NOTE: remove after gen2brain/raylib-go/pull/438 got merge
// MatrixDecompose - Decompose a transformation matrix into its rotational, translational and scaling components
func MatrixDecompose(mat rl.Matrix, translation *rl.Vector3, rotation *rl.Quaternion, scale *rl.Vector3) {
	// Extract translation.
	translation.X = mat.M12
	translation.Y = mat.M13
	translation.Z = mat.M14

	// Extract upper-left for determinant computation
	a := mat.M0
	b := mat.M4
	c := mat.M8
	d := mat.M1
	e := mat.M5
	f := mat.M9
	g := mat.M2
	h := mat.M6
	i := mat.M10
	A := e*i - f*h
	B := f*g - d*i
	C := d*h - e*g

	// Extract scale
	det := a*A + b*B + c*C
	abc := rl.NewVector3(a, b, c)
	def := rl.NewVector3(d, e, f)
	ghi := rl.NewVector3(g, h, i)

	scalex := rl.Vector3Length(abc)
	scaley := rl.Vector3Length(def)
	scalez := rl.Vector3Length(ghi)
	s := rl.NewVector3(scalex, scaley, scalez)

	if det < 0 {
		s = rl.Vector3Negate(s)
	}

	*scale = s

	// Remove scale from the matrix if it is not close to zero
	clone := mat
	if !rl.FloatEquals(det, 0) {
		clone.M0 /= s.X
		clone.M5 /= s.Y
		clone.M10 /= s.Z

		// Extract rotation
		*rotation = rl.QuaternionFromMatrix(clone)
	} else {
		// Set to identity if close to zero
		*rotation = rl.QuaternionIdentity()
	}
}
