package sound

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/mcbalaam/ebitter/pkg/embedfs"
)

type SoundPlayer struct {
	mu      sync.Mutex
	context *audio.Context
	buffers map[string][]byte
	bgm     *audio.Player
}

func NewSoundPlayer(sampleRate int) (*SoundPlayer, error) {
	return &SoundPlayer{
		context: audio.NewContext(sampleRate),
		buffers: make(map[string][]byte),
	}, nil
}

func (s *SoundPlayer) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.bgm != nil {
		s.bgm.Close()
		s.bgm = nil
	}
}

func (s *SoundPlayer) RegisterNewSound(path string, name string) error {
	f, err := embedfs.FS.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	stream, err := wav.Decode(s.context, f)
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

	player := audio.NewPlayerFromBytes(s.context, data)
	player.SetVolume(math.Max(0, math.Min(1, volume)))
	player.Play()
	return nil
}

func (s *SoundPlayer) PlayVariable(name string, volume float64, _ float64) error {
	return s.PlaySound(name, volume)
}

func (s *SoundPlayer) PlayBackground(name string, volume float64) error {
	s.mu.Lock()
	data, ok := s.buffers[name]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("no sound: %s", name)
	}
	if s.bgm != nil {
		s.bgm.Close()
		s.bgm = nil
	}
	s.mu.Unlock()

	reader := bytes.NewReader(data)
	loop := audio.NewInfiniteLoop(reader, int64(len(data)))
	player, err := audio.NewPlayer(s.context, loop)
	if err != nil {
		return fmt.Errorf("new player: %w", err)
	}
	player.SetVolume(math.Max(0, math.Min(1, volume)))
	player.Play()

	s.mu.Lock()
	s.bgm = player
	s.mu.Unlock()
	return nil
}
