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

	drawCenteredFitPixelText(card, "UPSKILL RUCOY", 512, 150, 500, 5, gold)
	drawCenteredFitPixelText(card, fmt.Sprintf("%d -> %d", data.FromSkill, data.ToSkill), panelCenterX, 365, 450, 5, green)

	timeText := compactRucoyDuration(data.EstimatedTime)
	drawCenteredFitPixelText(card, strings.ToUpper(data.Options.Vocation), panelCenterX, 595, 500, 4, gold)
	drawCenteredFitPixelText(card, "TEMPO "+timeText, panelCenterX, 650, 500, 4, cream)
	if data.DailyHours > 0 {
		drawCenteredFitPixelText(card, compactRucoyDailyDuration(data.EstimatedTime, data.DailyHours), panelCenterX, 705, 500, 2, blue)
	}

	drawCenteredFitPixelText(card, "MANA "+formatRucoyCardNumber(data.ManaEstimate.TotalMana), panelCenterX, 865, 530, 4, blue)
	drawCenteredFitPixelText(card, fmt.Sprintf("POTIONS %s - %s", formatRucoyCardNumber(data.ManaEstimate.MinPotions), formatRucoyCardNumber(data.ManaEstimate.MaxPotions)), panelCenterX, 925, 530, 4, cream)

	drawCenteredFitPixelText(card, "GOLD", panelCenterX, 1180, 620, 4, gold)
	drawCenteredFitPixelText(card, fmt.Sprintf("%s - %s", formatRucoyCardNumber(data.ManaEstimate.MinCost), formatRucoyCardNumber(data.ManaEstimate.MaxCost)), panelCenterX, 1245, 620, 3, cream)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, card, &jpeg.Options{Quality: 92}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
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
	width := pixelTextWidth(text, scale)
	drawPixelText(dst, text, centerX-width/2, y, scale, clr)
}

func drawCenteredFitPixelText(dst *image.RGBA, text string, centerX int, y int, maxWidth int, scale int, clr color.RGBA) {
	for scale > 1 && pixelTextWidth(text, scale) > maxWidth {
		scale--
	}
	drawCenteredPixelText(dst, text, centerX, y, scale, clr)
}

func drawFitPixelText(dst *image.RGBA, text string, x int, y int, maxWidth int, scale int, clr color.RGBA) {
	for scale > 1 && pixelTextWidth(text, scale) > maxWidth {
		scale--
	}
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
