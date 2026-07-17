package components

import "math"

func TransformPoint(p Vec, sx, sy, rot, tx, ty float64) Vec {
	x, y := p.X*sx, p.Y*sy
	c, s := math.Cos(rot), math.Sin(rot)
	rx := x*c - y*s
	ry := x*s + y*c
	return Vec{rx + tx, ry + ty}
}

func projectPoly(axis Vec, verts []Vec) (min, max float64) {
	min, max = math.Inf(1), math.Inf(-1)
	for _, v := range verts {
		proj := v.X*axis.X + v.Y*axis.Y
		if proj < min {
			min = proj
		}
		if proj > max {
			max = proj
		}
	}
	return
}

func overlapOnAxis(aMin, aMax, bMin, bMax float64) bool {
	return !(aMax < bMin || bMax < aMin)
}

func polygonsIntersect(a, b []Vec) bool {
	checkAxes := func(verts []Vec) bool {
		n := len(verts)
		for i := 0; i < n; i++ {
			j := (i + 1) % n
			ex := verts[j].X - verts[i].X
			ey := verts[j].Y - verts[i].Y
			axis := Vec{-ey, ex}
			aMin, aMax := projectPoly(axis, a)
			bMin, bMax := projectPoly(axis, b)
			if !overlapOnAxis(aMin, aMax, bMin, bMax) {
				return false
			}
		}
		return true
	}
	return checkAxes(a) && checkAxes(b)
}

func RectToPoly(width, height float64) []Vec {
	return []Vec{
		{0, 0},
		{width, 0},
		{width, height},
		{0, height},
	}
}

func TransformPoly(verts []Vec, sx, sy, rot, tx, ty float64) []Vec {
	out := make([]Vec, len(verts))
	for i, v := range verts {
		out[i] = TransformPoint(v, sx, sy, rot, tx, ty)
	}
	return out
}
