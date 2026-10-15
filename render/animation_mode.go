package render

// AnimationMode selects how an icon state advances through its frames.
type AnimationMode int

const (
	// AnimationModeOnce plays forward and stops on the last frame.
	AnimationModeOnce AnimationMode = iota
	// AnimationModeLoop plays forward and wraps to the first frame.
	AnimationModeLoop
	// AnimationModeReverse plays backward and wraps to the last frame.
	AnimationModeReverse
	// AnimationModePingPong plays forward then backward, repeating.
	AnimationModePingPong
	// AnimationModePingPongReverse plays backward then forward, repeating.
	AnimationModePingPongReverse
)

// animationModeFromAseprite maps an Aseprite frameTag direction to a mode.
func animationModeFromAseprite(direction string) AnimationMode {
	switch direction {
	case "forward", "normal", "fwd":
		return AnimationModeLoop
	case "reverse", "backward", "rev":
		return AnimationModeReverse
	case "pingpong", "ping-pong":
		return AnimationModePingPong
	case "pingpong_reverse", "pingpong-reverse", "pingpongreverse":
		return AnimationModePingPongReverse
	default:
		return AnimationModeOnce
	}
}
