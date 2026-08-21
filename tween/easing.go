package tween

import "math"

type EasingFunc func(t float64) float64

func Linear(t float64) float64 { return t }

func EaseInQuad(t float64) float64 { return t * t }

func EaseOutQuad(t float64) float64 { return t * (2 - t) }

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func EaseInCubic(t float64) float64 { return t * t * t }

func EaseOutCubic(t float64) float64 {
	t -= 1
	return t*t*t + 1
}

func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t -= 1
	return 4*t*t*t + 1
}

func EaseInQuart(t float64) float64 { return t * t * t * t }

func EaseOutQuart(t float64) float64 {
	t -= 1
	return -(t*t*t*t - 1)
}

func EaseInOutQuart(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	t -= 1
	return -(8*t*t*t*t - 1)
}

func EaseInQuint(t float64) float64 { return t * t * t * t * t }

func EaseOutQuint(t float64) float64 {
	t -= 1
	return t*t*t*t*t + 1
}

func EaseInOutQuint(t float64) float64 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	}
	t -= 1
	return 16*t*t*t*t*t + 1
}

func EaseInSine(t float64) float64 { return 1 - math.Cos(t*math.Pi/2) }

func EaseOutSine(t float64) float64 { return math.Sin(t * math.Pi / 2) }

func EaseInOutSine(t float64) float64 { return -(math.Cos(math.Pi*t) - 1) / 2 }

func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func EaseInOutExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return math.Pow(2, 10*(2*t-1)) / 2
	}
	return (2 - math.Pow(2, -10*(2*t-1))) / 2
}

func EaseInCirc(t float64) float64 { return 1 - math.Sqrt(1-t*t) }

func EaseOutCirc(t float64) float64 { return math.Sqrt(1 - (t-1)*(t-1)) }

func EaseInOutCirc(t float64) float64 {
	if t < 0.5 {
		return (1 - math.Sqrt(1-4*t*t)) / 2
	}
	return (math.Sqrt(1-(2*t-2)*(2*t-2)) + 1) / 2
}

func EaseInElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.075)*2*math.Pi/0.3)
}

func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t-0.075)*2*math.Pi/0.3) + 1
}

func EaseInOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	t *= 2
	if t < 1 {
		return -0.5 * math.Pow(2, 10*(t-1)) * math.Sin((t-1.075)*2*math.Pi/0.3)
	}
	t -= 1
	return 0.5*math.Pow(2, -10*t)*math.Sin((t-0.075)*2*math.Pi/0.3) + 1
}

func EaseInBack(t float64) float64 {
	const s = 1.70158
	return t * t * ((s+1)*t - s)
}

func EaseOutBack(t float64) float64 {
	const s = 1.70158
	t -= 1
	return t*t*((s+1)*t+s) + 1
}

func EaseInOutBack(t float64) float64 {
	const s = 1.70158 * 1.525
	t *= 2
	if t < 1 {
		return 0.5 * (t * t * ((s+1)*t - s))
	}
	t -= 2
	return 0.5 * (t*t*((s+1)*t+s) + 2)
}

func EaseOutBounce(t float64) float64 {
	if t < 1/2.75 {
		return 7.5625 * t * t
	} else if t < 2/2.75 {
		t -= 1.5 / 2.75
		return 7.5625*t*t + 0.75
	} else if t < 2.5/2.75 {
		t -= 2.25 / 2.75
		return 7.5625*t*t + 0.9375
	}
	t -= 2.625 / 2.75
	return 7.5625*t*t + 0.984375
}

func EaseInBounce(t float64) float64 { return 1 - EaseOutBounce(1-t) }

func EaseInOutBounce(t float64) float64 {
	if t < 0.5 {
		return (1 - EaseOutBounce(1-2*t)) / 2
	}
	return (1 + EaseOutBounce(2*t-1)) / 2
}
