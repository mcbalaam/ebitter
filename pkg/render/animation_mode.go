package render

type AnimationMode int

const (
	AnimationModeOnce     AnimationMode = iota
	AnimationModeLoop     AnimationMode = iota
	AnimationModePingPong AnimationMode = iota
)

var animationModeToString = [...]string{
	AnimationModeOnce:     "once",
	AnimationModeLoop:     "loop",
	AnimationModePingPong: "pingpong",
}
