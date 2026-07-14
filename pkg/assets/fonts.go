package assets

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

type FrameTag struct {
	Name      string `json:"name"`
	From      int    `json:"from"`
	To        int    `json:"to"`
	Direction string `json:"direction"`
	Color     string `json:"color"`
}

type FrameData struct {
	Frame struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	} `json:"frame"`
	Rotated          bool `json:"rotated"`
	Trimmed          bool `json:"trimmed"`
	SpriteSourceSize struct {
		X int `json:"x"`
		Y int `json:"y"`
		W int `json:"w"`
		H int `json:"h"`
	} `json:"spriteSourceSize"`
	SourceSize struct {
		W int `json:"w"`
		H int `json:"h"`
	} `json:"sourceSize"`
	Duration int `json:"duration"`
}

type MetaData struct {
	App     string `json:"app"`
	Version string `json:"version"`
	Image   string `json:"image"`
	Format  string `json:"format"`
	Size    struct {
		W int `json:"w"`
		H int `json:"h"`
	} `json:"size"`
	Scale     string        `json:"scale"`
	FrameTags []FrameTag    `json:"frameTags"`
	Layers    []interface{} `json:"layers"`
	Slices    []interface{} `json:"slices"`
}

type SpriteSheet struct {
	Frames map[string]FrameData `json:"frames"`
	Meta   MetaData             `json:"meta"`
}

type GlyphFrame struct {
	Index   int
	X       int
	Y       int
	Width   int
	Height  int
	OffsetY int
	Advance int
	Image   image.Image
}

func alreadyProcessed(dir, name string) bool {
	png := filepath.Join(dir, name+".png")
	json := filepath.Join(dir, name+".json")
	if _, err := os.Stat(png); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(json); os.IsNotExist(err) {
		return false
	}
	return true
}

func ProcessFonts() {
	outputDir := "media/sprites/"
	os.MkdirAll(outputDir, 0755)

	files, err := os.ReadDir("media/fonts")
	if err != nil {
		fmt.Printf("Error reading folder: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	maxWorkers := 4
	sem := make(chan struct{}, maxWorkers)

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".ttf") {
			continue
		}

		fontPath := filepath.Join("media/fonts", file.Name())
		fontName := strings.TrimSuffix(file.Name(), ".ttf")
		fontOutputDir := filepath.Join(outputDir, fontName)

		if alreadyProcessed(fontOutputDir, fontName) {
			fmt.Printf("Font already processed: %s (skipping)\n", fontName)
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(fileName, path, name, outDir string) {
			defer wg.Done()
			defer func() { <-sem }()

			fmt.Printf("Preprocessing font: %s...\n", fileName)

			os.MkdirAll(outDir, 0755)

			if err := processTTFFont(path, name, outDir); err != nil {
				fmt.Printf("Error processing %s: %v\n", fileName, err)
				return
			}

			fmt.Printf("Font processed: %s\n", name)
		}(file.Name(), fontPath, fontName, fontOutputDir)
	}

	wg.Wait()
}

const (
	FontSizePt = 32.0
	TargetDPI  = 288.0
)

func processTTFFont(fontPath, fontName, outputDir string) error {
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return err
	}

	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return err
	}

	fontSize := float64(FontSizePt)

	scale := TargetDPI / 72.0

	fontHeight := int(fontSize * scale * 1.4)
	baseline := int(fontSize * scale * 1.1)

	glyphFrames := []GlyphFrame{}

	ranges := []struct{ start, end int }{
		{32, 126},        // ASCII
		{0x0100, 0x017F}, // latin extended
		{0x0400, 0x044F}, // cyrrylic
	}

	for _, r := range ranges {
		for i := r.start; i <= r.end; i++ {
			img, err := renderGlyph(ttfFont, rune(i), fontSize, TargetDPI, fontHeight, baseline)
			if err != nil || img == nil {
				continue
			}

			if img.Bounds().Dx() > 0 {
				glyphFrames = append(glyphFrames, GlyphFrame{
					Index:  i,
					Width:  img.Bounds().Dx(),
					Height: fontHeight,
					Image:  img,
				})
			}
		}
	}

	if len(glyphFrames) == 0 {
		return fmt.Errorf("no glyphs found")
	}

	spriteImg, frames := arrangeGlyphs(glyphFrames)

	pngPath := filepath.Join(outputDir, fontName+".png")
	pngFile, err := os.Create(pngPath)
	if err != nil {
		return err
	}
	defer pngFile.Close()

	err = png.Encode(pngFile, spriteImg)
	if err != nil {
		return err
	}

	config := generateConfig(fontName, frames)

	jsonPath := filepath.Join(outputDir, fontName+".json")
	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(config)
	if err != nil {
		return err
	}

	return nil
}

func renderGlyph(ttfFont *truetype.Font, ch rune, fontSize float64, dpi float64, fontHeight int, ascent int) (image.Image, error) {
	canvasWidth := int(fontSize * 4)
	img := image.NewRGBA(image.Rect(0, 0, canvasWidth, fontHeight))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 0}), image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(ttfFont)
	c.SetFontSize(24)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.White))

	startX := 10
	pt := fixed.Point26_6{X: fixed.Int26_6(startX * 64), Y: fixed.Int26_6(ascent * 64)}

	_, err := c.DrawString(string(ch), pt)
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a >= 32768 {
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, 255})
			} else {
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	minX, maxX := canvasWidth, 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
		}
	}

	if minX == canvasWidth {
		emptyImg := image.NewRGBA(image.Rect(0, 0, 1, fontHeight))
		return emptyImg, nil
	}

	croppedBounds := image.Rect(minX, 0, maxX+1, fontHeight)
	croppedImg := image.NewRGBA(image.Rect(0, 0, croppedBounds.Dx(), croppedBounds.Dy()))
	draw.Draw(croppedImg, croppedImg.Bounds(), img, croppedBounds.Min, draw.Src)

	return croppedImg, nil
}

func arrangeGlyphs(glyphFrames []GlyphFrame) (image.Image, []GlyphFrame) {
	cols := 16
	rows := (len(glyphFrames) + cols - 1) / cols

	maxWidth := 0
	maxHeight := 0

	for _, gf := range glyphFrames {
		if gf.Width > maxWidth {
			maxWidth = gf.Width
		}
		if gf.Height > maxHeight {
			maxHeight = gf.Height
		}
	}

	cellWidth := maxWidth + 2
	cellHeight := maxHeight + 2

	spriteWidth := cols * cellWidth
	spriteHeight := rows * cellHeight

	spriteImg := image.NewRGBA(image.Rect(0, 0, spriteWidth, spriteHeight))
	draw.Draw(spriteImg, spriteImg.Bounds(), image.NewUniform(color.RGBA{0, 0, 0, 0}), image.Point{}, draw.Src)

	frames := make([]GlyphFrame, len(glyphFrames))
	for i, gf := range glyphFrames {
		col := i % cols
		row := i / cols

		cellX := col * cellWidth
		cellY := row * cellHeight

		drawX := cellX + 1
		drawY := cellY + 1

		draw.Draw(spriteImg, image.Rect(drawX, drawY, drawX+gf.Width, drawY+gf.Height),
			gf.Image, image.Point{}, draw.Over)

		frames[i] = GlyphFrame{
			Index:  gf.Index,
			X:      drawX,
			Y:      drawY,
			Width:  gf.Width,
			Height: gf.Height,
			Image:  gf.Image,
		}
	}

	return spriteImg, frames
}

func generateConfig(fontName string, frames []GlyphFrame) SpriteSheet {
	config := SpriteSheet{
		Frames: make(map[string]FrameData),
	}

	for i, frame := range frames {
		frameName := fmt.Sprintf("%s %d.aseprite", fontName, i)

		fd := FrameData{
			Rotated:  false,
			Trimmed:  false,
			Duration: 200,
		}

		fd.Frame.X = frame.X
		fd.Frame.Y = frame.Y
		fd.Frame.W = frame.Width
		fd.Frame.H = frame.Height

		fd.SpriteSourceSize.X = 0
		fd.SpriteSourceSize.Y = 0
		fd.SpriteSourceSize.W = frame.Width
		fd.SpriteSourceSize.H = frame.Height

		fd.SourceSize.W = frame.Width
		fd.SourceSize.H = frame.Height

		config.Frames[frameName] = fd
	}

	config.Meta.FrameTags = make([]FrameTag, len(frames))
	for i, frame := range frames {
		config.Meta.FrameTags[i] = FrameTag{
			Name:      string(rune(frame.Index)),
			From:      i,
			To:        i,
			Direction: "forward",
			Color:     "#000000ff",
		}
	}

	config.Meta.App = "http://www.aseprite.org/"
	config.Meta.Version = "1.3.17.2-dev"
	config.Meta.Image = fontName + ".png"
	config.Meta.Format = "RGBA8888"

	maxX, maxY := 0, 0
	for _, frame := range frames {
		if frame.X+frame.Width > maxX {
			maxX = frame.X + frame.Width
		}
		if frame.Y+frame.Height > maxY {
			maxY = frame.Y + frame.Height
		}
	}

	config.Meta.Size.W = maxX
	config.Meta.Size.H = maxY
	config.Meta.Scale = "1"
	config.Meta.Layers = []interface{}{
		map[string]interface{}{
			"name":      "Layer 1",
			"opacity":   255,
			"blendMode": "normal",
		},
	}
	config.Meta.Slices = []interface{}{}

	return config
}
