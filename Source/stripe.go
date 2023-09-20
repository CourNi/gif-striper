package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"math"
	"os"
	"quant/median"
	"resize"
	"time"
)

type Settings struct {
	ImageName         map[string]string `json:"images"`
	Offset            []int             `json:"offsets"`
	InterpolationType int               `json:"interpolation"`
	Watermark         bool              `json:"watermark"`
	Quantization      bool              `json:"quantization"`
}

func TimeTrack(start time.Time) {
	elapsed := time.Since(start)
	fmt.Printf("%s\n", elapsed)
}

// Check file exists
func Check(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("ERROR: Not found %s\n", filename)
		fmt.Print("Press ENTER to close...")
		reader.ReadString('\n')
		os.Exit(0)
	}
	file.Close()
}

func CreatePalette(images []image.Image) []color.Color {
	defer TimeTrack(time.Now())
	iLen := len(images)
	fmt.Printf("Palette quantization process - total %d images!\n", iLen)
	maxColor := int64(math.Floor(254 / float64(iLen)))
	var pallete []color.Color
	pallete = append(pallete, image.Transparent)
	pallete = append(pallete, image.White)
	for index, img := range images {
		fmt.Printf("Image %d median...", index+1)
		var q draw.Quantizer = median.Quantizer(maxColor)
		pal := q.Quantize(make(color.Palette, 0, maxColor), img)
		pallete = append(pallete, pal...)
		fmt.Println("Done")
	}
	fmt.Print("Quanization finished in ")
	return pallete
}

func LoadPNG(path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("ERROR: Mask unavailable")
		reader.ReadString('\n')
		os.Exit(0)
	}
	defer file.Close()

	imageData, _ := png.Decode(file)
	return imageData
}

func LoadGIF(filename string) *gif.GIF {
	file, err := os.Open(filename)
	if err != nil {
		os.Exit(0)
	}
	defer file.Close()

	imageData, _ := gif.DecodeAll(file)
	return imageData
}

func MinOf(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

func GetInterpolation(i int) resize.InterpolationFunction {
	switch i {
	case 0:
		return resize.NearestNeighbor
	case 1:
		return resize.Bilinear
	case 2:
		return resize.Bicubic
	case 3:
		return resize.Lanczos3
	default:
		return resize.NearestNeighbor
	}
}

func Draw() {
	defer TimeTrack(time.Now())

	//json
	data, setErr := os.ReadFile("settings.json")
	props := Settings{}
	if setErr == nil {
		err := json.Unmarshal(data, &props)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("ERROR: Settings file not found")
		reader.ReadString('\n')
		os.Exit(0)
	}

	//get files
	var masks []image.Image
	var filenames = []string{
		props.ImageName["monday"], props.ImageName["tuesday"], props.ImageName["wednesday"], props.ImageName["thursday"],
		props.ImageName["friday"], props.ImageName["saturday"], props.ImageName["sunday"],
	}

	for _, name := range filenames {
		Check(name)
	}
	fmt.Println("All source images found!")

	//load masks
	masks = append(masks, LoadPNG("masks/maskL"))
	masks = append(masks, LoadPNG("masks/maskC"))
	masks = append(masks, LoadPNG("masks/maskR"))
	masks = append(masks, LoadPNG("masks/sep"))

	//draw
	const w, h int = 1750, 244
	stripe := image.NewPaletted(image.Rect(0, 0, w, h), palette.Plan9)

	const disposalValue byte = 0x02
	const delay int = 1
	var images []*image.Paletted
	var delays []int
	var disposal []byte

	mon := LoadGIF(filenames[0])
	tue := LoadGIF(filenames[1])
	wed := LoadGIF(filenames[2])
	thu := LoadGIF(filenames[3])
	fri := LoadGIF(filenames[4])
	sat := LoadGIF(filenames[5])
	sun := LoadGIF(filenames[6])

	var pal []color.Color
	if props.Quantization {
		pal = CreatePalette(append([]image.Image{}, mon.Image[0], tue.Image[0], wed.Image[0], thu.Image[0], fri.Image[0], sat.Image[0], sun.Image[0]))
	} else {
		pal = palette.Plan9
	}

	frameCount := MinOf(len(mon.Image), len(tue.Image), len(wed.Image), len(thu.Image), len(fri.Image), len(sat.Image), len(sun.Image))
	fmt.Printf("Total Frames: %d\n", frameCount)

	r := GetInterpolation(props.InterpolationType)
	for i := 0; i < frameCount; i++ {
		fmt.Printf("Frame %d", i)
		stripe = image.NewPaletted(image.Rect(0, 0, w, h), pal)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(-15+props.Offset[0], -10, 265+props.Offset[0], h+10), resize.Resize(280, 264, mon.Image[i], r), image.Point{0, 0}, masks[0], image.Point{10 + props.Offset[0], 0}, draw.Over)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(235+props.Offset[1], -10, 515+props.Offset[1], h+10), resize.Resize(280, 264, tue.Image[i], r), image.Point{0, 0}, masks[1], image.Point{10 + props.Offset[1], 0}, draw.Over)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(485+props.Offset[2], -10, 765+props.Offset[2], h+10), resize.Resize(280, 264, wed.Image[i], r), image.Point{0, 0}, masks[1], image.Point{10 + props.Offset[2], 0}, draw.Over)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(735+props.Offset[3], -10, 1015+props.Offset[3], h+10), resize.Resize(280, 264, thu.Image[i], r), image.Point{0, 0}, masks[1], image.Point{10 + props.Offset[3], 0}, draw.Over)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(985+props.Offset[4], -10, 1265+props.Offset[4], h+10), resize.Resize(280, 264, fri.Image[i], r), image.Point{0, 0}, masks[1], image.Point{10 + props.Offset[4], 0}, draw.Over)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(1235+props.Offset[5], -10, 1515+props.Offset[5], h+10), resize.Resize(280, 264, sat.Image[i], r), image.Point{0, 0}, masks[1], image.Point{10 + props.Offset[5], 0}, draw.Over)
		fmt.Print(".")
		draw.DrawMask(stripe, image.Rect(1485+props.Offset[6], -10, 1765+props.Offset[6], h+10), resize.Resize(280, 264, sun.Image[i], r), image.Point{0, 0}, masks[2], image.Point{10 + props.Offset[6], 0}, draw.Over)
		fmt.Print(".")
		draw.Draw(stripe, image.Rect(0, 0, w, h), masks[3], image.Point{0, 0}, draw.Over)
		fmt.Print(".")
		images = append(images, stripe)
		delays = append(delays, delay)
		disposal = append(disposal, disposalValue)
		fmt.Println("Done")
	}

	//save
	f, err := os.OpenFile("stripe.gif", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image:    images,
		Delay:    delays,
		Disposal: disposal,
	})
	fmt.Print("GIF created for ")
}

func main() {
	Draw()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Window can be closed...")
	reader.ReadString('\n')
}
