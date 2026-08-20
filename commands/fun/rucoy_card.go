package fun

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	_ "image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

type RucoyUpskillCardData struct {
	FromSkill     int
	ToSkill       int
	EstimatedTime string
	DailyHours    int
	Options       RucoyUpskillOptions
	ManaEstimate  RucoyUpskillManaEstimate
}

func generateRucoyUpskillCard(data RucoyUpskillCardData) ([]byte, error) {
	template, err := loadRucoyUpskillCardTemplate()
	if err != nil {
		return nil, err
	}

	card := image.NewRGBA(template.Bounds())
	drawImage(card, template)

	cream := color.RGBA{R: 246, G: 232, B: 190, A: 255}
	gold := color.RGBA{R: 246, G: 196, B: 69, A: 255}
	green := color.RGBA{R: 139, G: 221, B: 59, A: 255}
	blue := color.RGBA{R: 93, G: 210, B: 255, A: 255}

	panelCenterX := 596

	drawCenteredPixelTextBlock(card, 512, 118, 228, []rucoyCardTextLine{
		{text: "UPSKILL RUCOY", maxWidth: 500, scale: 5, color: gold},
	}, 0)
	drawCenteredPixelTextBlock(card, panelCenterX, 332, 445, []rucoyCardTextLine{
		{text: fmt.Sprintf("%d -> %d", data.FromSkill, data.ToSkill), maxWidth: 450, scale: 5, color: green},
	}, 0)

	timeText := compactRucoyDuration(data.EstimatedTime)
	timeLines := []rucoyCardTextLine{
		{text: strings.ToUpper(data.Options.Vocation), maxWidth: 500, scale: 2, color: gold},
		{text: "TEMPO", maxWidth: 500, scale: 3, color: gold},
		{text: timeText, maxWidth: 500, scale: 3, color: cream},
	}
	if data.DailyHours > 0 {
		timeLines = append(timeLines, rucoyCardTextLine{text: compactRucoyDailyDuration(data.EstimatedTime, data.DailyHours), maxWidth: 500, scale: 2, color: blue})
	}
	drawCenteredPixelTextBlock(card, panelCenterX, 586, 782, timeLines, 6)

	drawCenteredPixelTextBlock(card, panelCenterX, 858, 1044, []rucoyCardTextLine{
		{text: "MANA", maxWidth: 530, scale: 2, color: blue},
		{text: formatRucoyCardNumber(data.ManaEstimate.TotalMana), maxWidth: 530, scale: 3, color: cream},
		{text: "POTIONS", maxWidth: 530, scale: 2, color: blue},
		{text: fmt.Sprintf("%s - %s", formatRucoyCardNumber(data.ManaEstimate.MinPotions), formatRucoyCardNumber(data.ManaEstimate.MaxPotions)), maxWidth: 530, scale: 3, color: cream},
	}, 6)

	drawCenteredPixelTextBlock(card, panelCenterX, 1100, 1322, []rucoyCardTextLine{
		{text: "GOLD", maxWidth: 620, scale: 3, color: gold},
		{text: fmt.Sprintf("%s - %s", formatRucoyCardNumber(data.ManaEstimate.MinCost), formatRucoyCardNumber(data.ManaEstimate.MaxCost)), maxWidth: 620, scale: 3, color: cream},
	}, 18)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, card, &jpeg.Options{Quality: 92}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

type rucoyCardTextLine struct {
	text     string
	maxWidth int
	scale    int
	color    color.RGBA
}

func loadRucoyUpskillCardTemplate() (image.Image, error) {
	paths := []string{
		filepath.Join("assets", "rucoy", "upskill-card-template.png"),
		filepath.Join("..", "..", "assets", "rucoy", "upskill-card-template.png"),
	}

	var file *os.File
	var err error
	for _, path := range paths {
		file, err = os.Open(path)
		if err == nil {
			break
		}
	}
	if file == nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

func drawImage(dst *image.RGBA, src image.Image) {
	bounds := src.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.Set(x, y, src.At(x, y))
		}
	}
}

func drawCenteredPixelText(dst *image.RGBA, text string, centerX int, y int, scale int, clr color.RGBA) {
	width := pixelTextVisualWidth(text, scale)
	drawPixelText(dst, text, centerX-width/2, y, scale, clr)
}

func drawCenteredFitPixelText(dst *image.RGBA, text string, centerX int, y int, maxWidth int, scale int, clr color.RGBA) {
	scale = fitPixelTextScale(text, maxWidth, scale)
	drawCenteredPixelText(dst, text, centerX, y, scale, clr)
}

func drawCenteredPixelTextBlock(dst *image.RGBA, centerX int, topY int, bottomY int, lines []rucoyCardTextLine, gap int) {
	if len(lines) == 0 || bottomY <= topY {
		return
	}

	scales := make([]int, len(lines))
	totalHeight := 0
	for index, line := range lines {
		scales[index] = fitPixelTextScale(line.text, line.maxWidth, line.scale)
		totalHeight += pixelTextVisualHeight(scales[index])
	}
	totalHeight += gap * (len(lines) - 1)

	y := topY + (bottomY-topY-totalHeight)/2
	for index, line := range lines {
		drawCenteredPixelText(dst, line.text, centerX, y, scales[index], line.color)
		y += pixelTextVisualHeight(scales[index]) + gap
	}
}

func fitPixelTextScale(text string, maxWidth int, scale int) int {
	for scale > 1 && pixelTextVisualWidth(text, scale) > maxWidth {
		scale--
	}
	return scale
}

func drawFitPixelText(dst *image.RGBA, text string, x int, y int, maxWidth int, scale int, clr color.RGBA) {
	scale = fitPixelTextScale(text, maxWidth, scale)
	drawPixelText(dst, text, x, y, scale, clr)
}

func drawPixelText(dst *image.RGBA, text string, x int, y int, scale int, clr color.RGBA) {
	drawPixelTextLayer(dst, text, x+scale, y+scale, scale, color.RGBA{A: 180})
	drawPixelTextLayer(dst, text, x, y, scale, clr)
}

func drawPixelTextLayer(dst *image.RGBA, text string, x int, y int, scale int, clr color.RGBA) {
	face := basicfont.Face7x13
	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	height := ascent + metrics.Descent.Ceil()
	width := font.MeasureString(face, text).Ceil()
	if width <= 0 || height <= 0 {
		return
	}

	mask := image.NewRGBA(image.Rect(0, 0, width, height))
	drawer := font.Drawer{
		Dst:  mask,
		Src:  image.NewUniform(clr),
		Face: face,
		Dot:  fixed.P(0, ascent),
	}
	drawer.DrawString(text)

	for sy := 0; sy < height; sy++ {
		for sx := 0; sx < width; sx++ {
			_, _, _, alpha := mask.At(sx, sy).RGBA()
			if alpha == 0 {
				continue
			}
			for oy := 0; oy < scale; oy++ {
				for ox := 0; ox < scale; ox++ {
					dst.Set(x+sx*scale+ox, y+sy*scale+oy, clr)
				}
			}
		}
	}
}

func pixelTextWidth(text string, scale int) int {
	return font.MeasureString(basicfont.Face7x13, text).Ceil() * scale
}

func pixelTextVisualWidth(text string, scale int) int {
	return pixelTextWidth(text, scale) + scale
}

func pixelTextHeight(scale int) int {
	metrics := basicfont.Face7x13.Metrics()
	return (metrics.Ascent.Ceil() + metrics.Descent.Ceil()) * scale
}

func pixelTextVisualHeight(scale int) int {
	return pixelTextHeight(scale) + scale
}

func compactRucoyDuration(value string) string {
	totalMinutes, ok := parseRucoyFormattedDurationMinutes(value)
	if !ok {
		return strings.ToUpper(value)
	}
	return compactRucoyMinutes(totalMinutes)
}

func compactRucoyDailyDuration(value string, dailyHours int) string {
	totalMinutes, ok := parseRucoyFormattedDurationMinutes(value)
	if !ok || dailyHours <= 0 {
		return ""
	}

	dailyMinutes := int64(dailyHours * 60)
	days := totalMinutes / dailyMinutes
	remaining := totalMinutes % dailyMinutes
	return fmt.Sprintf("%dH/DIA: %dD %s", dailyHours, days, compactRucoyMinutes(remaining))
}

func compactRucoyMinutes(totalMinutes int64) string {
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dMIN", minutes)
	}
	if minutes == 0 {
		return fmt.Sprintf("%dH", hours)
	}
	return fmt.Sprintf("%dH %dMIN", hours, minutes)
}

func formatRucoyCardNumber(value int64) string {
	switch {
	case value >= 1000000:
		return strings.ReplaceAll(fmt.Sprintf("%.1fKK", float64(value)/1000000), ".0KK", "KK")
	case value >= 1000:
		return strings.ReplaceAll(fmt.Sprintf("%.0fK", float64(value)/1000), ".0K", "K")
	default:
		return formatRucoyNumber(value)
	}
}
