package assets

import (
	"encoding/json"
	"image"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type FrameMeta struct {
	Rect       image.Rectangle
	DurationMS int
}
type IconStateMeta struct {
	Name   string
	Frames []FrameMeta
	Mode   string
}

// internal shapes for reading Aseprite JSON outputs
type aseFrameRaw struct {
	Frame struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	} `json:"frame"`
	Duration int `json:"duration"`
}

type aseTagRaw struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
}

type aseMetaRaw struct {
	FrameTags []aseTagRaw `json:"frameTags"`
}

type aseJSON struct {
	Frames map[string]aseFrameRaw `json:"frames"`
	Meta   aseMetaRaw             `json:"meta"`
}

func ParseAsepriteJSON(data []byte) ([]IconStateMeta, error) {
	var a aseJSON
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}

	type keyed struct {
		key   string
		index int
		val   aseFrameRaw
	}
	var items []keyed

	re := regexp.MustCompile(`(\d+)(?:\D*$)`)
	for k, v := range a.Frames {
		idx := -1
		m := re.FindStringSubmatch(k)
		if len(m) > 1 {
			if n, err := strconv.Atoi(m[1]); err == nil {
				idx = n
			}
		}
		items = append(items, keyed{key: k, index: idx, val: v})
	}

	sort.SliceStable(items, func(i, j int) bool {
		aIdx, bIdx := items[i].index, items[j].index
		if aIdx >= 0 && bIdx >= 0 {
			return aIdx < bIdx
		}
		if aIdx >= 0 {
			return true
		}
		if bIdx >= 0 {
			return false
		}
		return strings.Compare(items[i].key, items[j].key) < 0
	})

	ordered := make([]aseFrameRaw, 0, len(items))
	for _, it := range items {
		ordered = append(ordered, it.val)
	}

	var states []IconStateMeta
	for _, tag := range a.Meta.FrameTags {
		from := tag.From
		to := tag.To
		if from < 0 {
			from = 0
		}
		if to >= len(ordered) {
			to = len(ordered) - 1
		}
		if from > to {
			continue
		}
		fs := make([]FrameMeta, 0, to-from+1)
		for i := from; i <= to; i++ {
			f := ordered[i]
			r := image.Rect(f.Frame.X, f.Frame.Y, f.Frame.X+f.Frame.W, f.Frame.Y+f.Frame.H)
			fs = append(fs, FrameMeta{
				Rect:       r,
				DurationMS: f.Duration,
			})
		}
		states = append(states, IconStateMeta{
			Name:   tag.Name,
			Frames: fs,
			Mode:   strings.ToLower(tag.Direction),
		})
	}

	return states, nil
}
