package rucoy

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
	template, err := loadRucoyUpskillCardTemplate(data.Options)
	if err != nil {
		return nil, err
	}

	card := image.NewRGBA(template.Bounds())
	drawImage(card, template)

	cream := color.RGBA{R: 246, G: 232, B: 190, A: 255}
	gold := color.RGBA{R: 246, G: 196, B: 69, A: 255}
	green := color.RGBA{R: 139, G: 221, B: 59, A: 255}
	blue := color.RGBA{R: 93, G: 210, B: 255, A: 255}

	titleCenterX := 512
	skillCenterX := 594
	timeCenterX := 600
	manaCenterX := 594
	goldCenterX := 612
	pallyGoldCenterX := 570

	drawCenteredFitPixelTextAtCenter(card, "UPSKILL RUCOY", titleCenterX, 204, 500, 5, gold)
	drawCenteredFitPixelTextAtCenter(card, fmt.Sprintf("%d -> %d", data.FromSkill, data.ToSkill), skillCenterX, 418, 450, 5, green)

	timeText := compactRucoyDuration(data.EstimatedTime)
	if data.ManaEstimate.TotalArrows > 0 {
		drawCenteredFitPixelTextAtCenter(card, "TEMPO", timeCenterX, 620, 500, 3, gold)
		drawCenteredFitPixelTextAtCenter(card, timeText, timeCenterX, 660, 500, 3, cream)
		if data.DailyHours > 0 {
			drawCenteredFitPixelTextAtCenter(card, compactRucoyDailyDuration(data.EstimatedTime, data.DailyHours), timeCenterX, 700, 500, 2, blue)
		}

		drawCenteredFitPixelTextAtCenter(card, "MANA "+formatRucoyCardNumber(data.ManaEstimate.TotalMana), manaCenterX, 850, 530, 3, blue)
		drawCenteredFitPixelTextAtCenter(card, fmt.Sprintf("POTIONS %s - %s", formatRucoyCardNumber(data.ManaEstimate.MinPotions), formatRucoyCardNumber(data.ManaEstimate.MaxPotions)), manaCenterX, 910, 530, 3, cream)

		drawCenteredFitPixelTextAtCenter(card, "FLECHAS "+formatRucoyCardNumber(data.ManaEstimate.TotalArrows), manaCenterX, 1060, 530, 3, blue)
		drawCenteredFitPixelTextAtCenter(card, "CUSTO "+formatRucoyCardNumber(data.ManaEstimate.ArrowCost), manaCenterX, 1120, 530, 3, cream)

		drawCenteredFitPixelTextAtCenter(card, "GOLD", pallyGoldCenterX, 1275, 500, 3, gold)
		drawCenteredFitPixelTextAtCenter(card, fmt.Sprintf("%s - %s", formatRucoyCardNumber(data.ManaEstimate.MinCost), formatRucoyCardNumber(data.ManaEstimate.MaxCost)), pallyGoldCenterX, 1332, 500, 3, cream)
	} else {
		drawCenteredFitPixelTextAtCenter(card, "TEMPO", timeCenterX, 664, 500, 3, gold)
		drawCenteredFitPixelTextAtCenter(card, timeText, timeCenterX, 708, 500, 3, cream)
		if data.DailyHours > 0 {
			drawCenteredFitPixelTextAtCenter(card, compactRucoyDailyDuration(data.EstimatedTime, data.DailyHours), timeCenterX, 750, 500, 2, blue)
		}

		drawCenteredFitPixelTextAtCenter(card, "MANA "+formatRucoyCardNumber(data.ManaEstimate.TotalMana), manaCenterX, 926, 530, 3, blue)
		drawCenteredFitPixelTextAtCenter(card, fmt.Sprintf("POTIONS %s - %s", formatRucoyCardNumber(data.ManaEstimate.MinPotions), formatRucoyCardNumber(data.ManaEstimate.MaxPotions)), manaCenterX, 986, 530, 3, cream)

		drawCenteredFitPixelTextAtCenter(card, "GOLD", goldCenterX, 1192, 620, 3, gold)
		drawCenteredFitPixelTextAtCenter(card, fmt.Sprintf("%s - %s", formatRucoyCardNumber(data.ManaEstimate.MinCost), formatRucoyCardNumber(data.ManaEstimate.MaxCost)), goldCenterX, 1248, 620, 3, cream)
	}

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

func loadRucoyUpskillCardTemplate(options RucoyUpskillOptions) (image.Image, error) {
	templateName := "upskill-card-template.png"
	if options.Vocation == "Pally" {
		templateName = "upskill-card-template-pally-arrows.png"
	} else if options.Vocation == "Mage" {
		templateName = "upskill-card-template-mage-fire.png"
	}

	paths := []string{
		filepath.Join("assets", "rucoy", templateName),
		filepath.Join("..", "..", "assets", "rucoy", templateName),
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
		return nil, fmt.Errorf("template %s nao encontrado em %s: %w", templateName, strings.Join(paths, ", "), err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler template %s: %w", templateName, err)
	}
	return img, nil
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

func drawCenteredFitPixelTextAtCenter(dst *image.RGBA, text string, centerX int, centerY int, maxWidth int, scale int, clr color.RGBA) {
	scale = fitPixelTextScale(text, maxWidth, scale)
	y := centerY - pixelTextVisualHeight(scale)/2
	drawCenteredPixelText(dst, text, centerX, y, scale, clr)
}

func drawCenteredPixelTextBlock(dst *image.RGBA, centerX int, topY int, bottomY int, lines []rucoyCardTextLine, gap int) {
	if len(lines) == 0 || bottomY <= topY {
		return
	}

	totalHeight, scales := fittedPixelTextBlockHeight(lines, gap)
	y := topY + (bottomY-topY-totalHeight)/2
	drawPixelTextBlock(dst, centerX, y, lines, scales, gap)
}

func drawCenteredPixelTextBlockAtCenter(dst *image.RGBA, centerX int, centerY int, lines []rucoyCardTextLine, gap int) {
	if len(lines) == 0 {
		return
	}

	totalHeight, scales := fittedPixelTextBlockHeight(lines, gap)
	y := centerY - totalHeight/2
	drawPixelTextBlock(dst, centerX, y, lines, scales, gap)
}

func fittedPixelTextBlockHeight(lines []rucoyCardTextLine, gap int) (int, []int) {
	scales := make([]int, len(lines))
	totalHeight := 0
	for index, line := range lines {
		scales[index] = fitPixelTextScale(line.text, line.maxWidth, line.scale)
		totalHeight += pixelTextVisualHeight(scales[index])
	}
	totalHeight += gap * (len(lines) - 1)
	return totalHeight, scales
}

func drawPixelTextBlock(dst *image.RGBA, centerX int, y int, lines []rucoyCardTextLine, scales []int, gap int) {
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
