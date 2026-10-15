package render

import (
	"testing"
	"time"
)

func testFrames(n int) []Frame {
	out := make([]Frame, n)
	for i := range out {
		out[i] = Frame{Time: time.Millisecond}
	}
	return out
}

func newTestIcon(mode AnimationMode, n int) *AnimatedIcon {
	st := &IconState{Name: "s", Frames: testFrames(n), Mode: mode}
	st.CurrentFrame = initialFrame(mode, n)
	st.CurrentFrameRef = &st.Frames[st.CurrentFrame]
	st.dir = initialDir(mode)
	return &AnimatedIcon{CurrentState: st}
}

func TestAnimationModeFromAseprite(t *testing.T) {
	cases := map[string]AnimationMode{
		"forward":          AnimationModeLoop,
		"normal":           AnimationModeLoop,
		"reverse":          AnimationModeReverse,
		"backward":         AnimationModeReverse,
		"pingpong":         AnimationModePingPong,
		"pingpong_reverse": AnimationModePingPongReverse,
		"":                 AnimationModeOnce,
		"weird":            AnimationModeOnce,
	}
	for dir, want := range cases {
		if got := animationModeFromAseprite(dir); got != want {
			t.Errorf("%q -> %v, want %v", dir, got, want)
		}
	}
}

func TestLoopStepsForward(t *testing.T) {
	icon := newTestIcon(AnimationModeLoop, 3)
	icon.Update(time.Millisecond)
	if got := icon.CurrentState.CurrentFrame; got != 1 {
		t.Fatalf("frame = %d, want 1", got)
	}
}

func TestReverseStepsBackward(t *testing.T) {
	icon := newTestIcon(AnimationModeReverse, 3)
	icon.Update(time.Millisecond)
	if got := icon.CurrentState.CurrentFrame; got != 2 {
		t.Fatalf("frame = %d, want 2", got)
	}
}

func TestPingPongReverseStartsAtLastAndGoesBack(t *testing.T) {
	icon := newTestIcon(AnimationModePingPongReverse, 3)
	if got := icon.CurrentState.CurrentFrame; got != 2 {
		t.Fatalf("initial frame = %d, want 2", got)
	}
	icon.Update(time.Millisecond)
	if got := icon.CurrentState.CurrentFrame; got != 1 {
		t.Fatalf("frame = %d, want 1", got)
	}
}
