package sound

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/mcbalaam/ebitter/assetfs"
)

// Default pitch multiplier bounds applied by PlayVariable. The lower bound is
// an octave down, the upper an octave up.
const (
	defaultMinPitch = 0.5
	defaultMaxPitch = 2.0
)

// Ebitengine audio buffers are signed 16-bit little-endian stereo PCM, so a
// single frame is two channels worth of 16-bit samples.
const bytesPerFrame = 4

type SoundPlayer struct {
	mu       sync.Mutex
	context  *audio.Context
	buffers  map[string][]byte
	bgm      *audio.Player
	minPitch float64
	maxPitch float64
	// FS is the filesystem sounds are read from. Nil falls back to assetfs.FS.
	FS fs.FS
}

func (s *SoundPlayer) fsys() fs.FS {
	if s.FS != nil {
		return s.FS
	}
	return assetfs.FS
}

func NewSoundPlayer(sampleRate int) (*SoundPlayer, error) {
	if sampleRate <= 0 {
		return nil, fmt.Errorf("invalid sample rate: %d", sampleRate)
	}
	return &SoundPlayer{
		context:  audio.NewContext(sampleRate),
		buffers:  make(map[string][]byte),
		minPitch: defaultMinPitch,
		maxPitch: defaultMaxPitch,
	}, nil
}

// SetPitchRange restricts the pitch multipliers accepted by PlayVariable to
// [min, max]. A non-positive min or a max below min restores the defaults
// (0.5 and 2.0).
func (s *SoundPlayer) SetPitchRange(min, max float64) {
	if min <= 0 || max < min {
		min, max = defaultMinPitch, defaultMaxPitch
	}
	s.mu.Lock()
	s.minPitch, s.maxPitch = min, max
	s.mu.Unlock()
}

func (s *SoundPlayer) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bgm != nil {
		s.bgm.Close()
		s.bgm = nil
	}
}

// RegisterNewSound decodes an OGG (Vorbis) file and stores it under name for
// reuse. The stream is resampled to the player's sample rate.
func (s *SoundPlayer) RegisterNewSound(path string, name string) error {
	f, err := s.fsys().Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	stream, err := vorbis.Decode(s.context, f)
	if err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}

	data, err := io.ReadAll(stream)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.buffers[name]; exists {
		return fmt.Errorf("sound exists: %s", name)
	}
	s.buffers[name] = data
	return nil
}

func (s *SoundPlayer) PlaySound(name string, volume float64) error {
	s.mu.Lock()
	data, ok := s.buffers[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no sound: %s", name)
	}

	s.play(data, volume)
	return nil
}

// PlayVariable plays a registered sound at the given volume with a pitch
// multiplier. A pitch of 1.0 is unmodified; greater values raise the pitch
// (and shorten the sound), lower values deepen it. The pitch is clamped to the
// range configured with SetPitchRange.
//
// Ebitengine has no playback-rate control, so the PCM data is resampled with
// linear interpolation before playback.
func (s *SoundPlayer) PlayVariable(name string, volume float64, pitch float64) error {
	s.mu.Lock()
	data, ok := s.buffers[name]
	minPitch, maxPitch := s.minPitch, s.maxPitch
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no sound: %s", name)
	}

	s.play(resamplePCM(data, clamp(pitch, minPitch, maxPitch)), volume)
	return nil
}

// play starts a one-shot player and closes it once playback finishes.
func (s *SoundPlayer) play(data []byte, volume float64) {
	player := s.context.NewPlayerFromBytes(data)
	player.SetVolume(clamp(volume, 0, 1))
	player.Play()
	go func() {
		for player.IsPlaying() {
			time.Sleep(50 * time.Millisecond)
		}
		player.Close()
	}()
}

// resamplePCM resamples interleaved 16-bit little-endian stereo PCM by a pitch
// factor using linear interpolation. pitch > 1 shortens the buffer (higher
// pitch), pitch < 1 lengthens it (lower pitch). A pitch of 1 returns the input
// unchanged. The input is never modified.
func resamplePCM(in []byte, pitch float64) []byte {
	if pitch == 1 || len(in) < bytesPerFrame*2 {
		return in
	}
	if pitch <= 0 {
		pitch = 1
	}

	inFrames := len(in) / bytesPerFrame
	outFrames := int(float64(inFrames) / pitch)
	if outFrames < 1 {
		outFrames = 1
	}

	out := make([]byte, outFrames*bytesPerFrame)
	last := inFrames - 1
	for i := 0; i < outFrames; i++ {
		src := float64(i) * pitch
		i0 := int(src)
		if i0 < 0 {
			i0 = 0
		} else if i0 > last {
			i0 = last
		}
		i1 := i0 + 1
		if i1 > last {
			i1 = last
		}
		frac := src - float64(i0)
		if frac < 0 {
			frac = 0
		}

		for ch := 0; ch < 2; ch++ {
			off := ch * 2
			a := float64(int16(binary.LittleEndian.Uint16(in[i0*bytesPerFrame+off:])))
			b := float64(int16(binary.LittleEndian.Uint16(in[i1*bytesPerFrame+off:])))
			binary.LittleEndian.PutUint16(out[i*bytesPerFrame+off:], uint16(int16(math.Round(a+(b-a)*frac))))
		}
	}
	return out
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (s *SoundPlayer) PlayBackground(name string, volume float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.buffers[name]
	if !ok {
		return fmt.Errorf("no sound: %s", name)
	}

	reader := bytes.NewReader(data)
	loop := audio.NewInfiniteLoop(reader, int64(len(data)))
	player, err := audio.NewPlayer(s.context, loop)
	if err != nil {
		return fmt.Errorf("new player: %w", err)
	}
	player.SetVolume(clamp(volume, 0, 1))
	player.Play()

	if s.bgm != nil {
		s.bgm.Close()
	}
	s.bgm = player
	return nil
}
