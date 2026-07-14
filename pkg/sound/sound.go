package sound

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

type SoundPlayer struct {
	Sounds     map[string]*Sound
	mu         sync.Mutex
	sampleRate beep.SampleRate
}

type Sound struct {
	Path string
}

func NewSoundPlayer(sampleRate beep.SampleRate) (*SoundPlayer, error) {
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/100))
	if err != nil {
		return nil, fmt.Errorf("speaker init: %w", err)
	}

	return &SoundPlayer{
		Sounds:     make(map[string]*Sound),
		sampleRate: sampleRate,
	}, nil
}

func (s *SoundPlayer) Shutdown() {
	speaker.Close()
}

func (s *SoundPlayer) RegisterNewSound(path string, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	_, _, err = wav.Decode(file)
	if err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Sounds[name] != nil {
		return fmt.Errorf("sound exists: %s", name)
	}

	s.Sounds[name] = &Sound{
		Path: path,
	}
	return nil
}

func (s *SoundPlayer) PlayBackground(name string, volume float64) error {
	s.mu.Lock()
	sound, ok := s.Sounds[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no sound: %s", name)
	}
	return sound.PlayLooped(volume, 1.0)
}

func (s *SoundPlayer) PlaySound(name string, volume float64) error {
	s.mu.Lock()
	sound, ok := s.Sounds[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no sound: %s", name)
	}
	return sound.Play(volume, 1.0)
}

func (s *SoundPlayer) PlayVariable(name string, volume float64, pitchVariation float64) error {
	s.mu.Lock()
	sound, ok := s.Sounds[name]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("no sound: %s", name)
	}

	pitch := 1.0 + (rand.Float64()-0.5)*pitchVariation
	return sound.Play(volume, pitch)
}

func (s *Sound) PlayLooped(volume float64, pitch float64) error {
	file, err := os.Open(s.Path)
	if err != nil {
		return fmt.Errorf("open %s: %w", s.Path, err)
	}

	streamer, _, err := wav.Decode(file)
	if err != nil {
		file.Close()
		return fmt.Errorf("decode %s: %w", s.Path, err)
	}

	looped := beep.Loop(-1, streamer)
	ctrl := &closableStreamer{
		Streamer: looped,
		file:     file,
	}

	if pitch != 1.0 {
		ctrl = &closableStreamer{
			Streamer: &pitchShifter{
				streamer: ctrl,
				pitch:    pitch,
				bufPos:   0,
			},
			file: file,
		}
	}

	if volume != 1.0 {
		ctrl = &closableStreamer{
			Streamer: &effects.Volume{
				Streamer: ctrl,
				Base:     2,
				Volume:   volume,
			},
			file: file,
		}
	}

	speaker.Play(ctrl)
	return nil
}

func (s *Sound) Play(volume float64, pitch float64) error {
	file, err := os.Open(s.Path)
	if err != nil {
		return fmt.Errorf("open %s: %w", s.Path, err)
	}

	streamer, _, err := wav.Decode(file)
	if err != nil {
		file.Close()
		return fmt.Errorf("decode %s: %w", s.Path, err)
	}

	ctrl := &closableStreamer{
		Streamer: streamer,
		file:     file,
	}

	if pitch != 1.0 {
		ctrl = &closableStreamer{
			Streamer: &pitchShifter{
				streamer: ctrl,
				pitch:    pitch,
				bufPos:   0,
			},
			file: file,
		}
	}

	if volume != 1.0 {
		ctrl = &closableStreamer{
			Streamer: &effects.Volume{
				Streamer: ctrl,
				Base:     2,
				Volume:   volume,
			},
			file: file,
		}
	}

	speaker.Play(ctrl)
	return nil
}

type closableStreamer struct {
	Streamer beep.Streamer
	file     io.Closer
}

func (cs *closableStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = cs.Streamer.Stream(samples)
	if !ok {
		cs.file.Close()
	}
	return
}

func (cs *closableStreamer) Err() error {
	if se, ok := cs.Streamer.(interface{ Err() error }); ok {
		return se.Err()
	}
	return nil
}

type pitchShifter struct {
	streamer beep.Streamer
	pitch    float64
	buf      [][2]float64
	bufPos   float64
	bufEnd   int
	bufFull  bool
	history  [][2]float64
}

func (ps *pitchShifter) Stream(samples [][2]float64) (n int, ok bool) {
	if ps.buf == nil {
		ps.buf = make([][2]float64, 4096)
		ps.history = make([][2]float64, 0, 4)
		ps.bufEnd = 0
		ps.bufFull = false
	}

	outIdx := 0

	for outIdx < len(samples) {
		if int(ps.bufPos) >= ps.bufEnd {
			if ps.bufEnd > 0 {
				start := ps.bufEnd - 3
				if start < 0 {
					start = 0
				}
				ps.history = append(ps.history[:0], ps.buf[start:ps.bufEnd]...)
			}

			readN, readOk := ps.streamer.Stream(ps.buf)
			if readN == 0 {
				return outIdx, false
			}

			ps.bufPos = 0
			ps.bufEnd = readN
			ps.bufFull = readOk

			continue
		}

		sample := ps.interpolateCubic(ps.bufPos)
		samples[outIdx] = sample
		outIdx++

		ps.bufPos += ps.pitch
	}

	return outIdx, true
}

func (ps *pitchShifter) interpolateCubic(pos float64) [2]float64 {
	idx := int(pos)
	frac := pos - float64(idx)

	p0 := ps.getSampleWithHistory(idx - 1)
	p1 := ps.getSampleWithHistory(idx)
	p2 := ps.getSampleWithHistory(idx + 1)
	p3 := ps.getSampleWithHistory(idx + 2)

	a0 := -0.5*p0[0] + 1.5*p1[0] - 1.5*p2[0] + 0.5*p3[0]
	a1 := -0.5*p0[1] + 1.5*p1[1] - 1.5*p2[1] + 0.5*p3[1]

	a2 := p0[0] - 2.5*p1[0] + 2.0*p2[0] - 0.5*p3[0]
	b2 := p0[1] - 2.5*p1[1] + 2.0*p2[1] - 0.5*p3[1]

	a3 := -0.5*p0[0] + 0.0*p1[0] + 0.5*p2[0] + 0.0*p3[0]
	b3 := -0.5*p0[1] + 0.0*p1[1] + 0.5*p2[1] + 0.0*p3[1]

	a4 := p1[0]
	b4 := p1[1]

	fracSq := frac * frac
	fracCube := fracSq * frac

	l := a0*fracCube + a2*fracSq + a3*frac + a4
	r := a1*fracCube + b2*fracSq + b3*frac + b4

	return [2]float64{l, r}
}

func (ps *pitchShifter) getSampleWithHistory(idx int) [2]float64 {
	if idx < 0 {
		histIdx := len(ps.history) + idx
		if histIdx >= 0 && histIdx < len(ps.history) {
			return ps.history[histIdx]
		}
		if ps.bufEnd > 0 {
			return ps.buf[0]
		}
		return [2]float64{0, 0}
	}

	if idx < ps.bufEnd {
		return ps.buf[idx]
	}

	return [2]float64{0, 0}
}

func (ps *pitchShifter) Err() error {
	return ps.streamer.Err()
}
