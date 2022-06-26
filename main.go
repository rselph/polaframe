package main

import (
	"flag"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/image/draw"
	"golang.org/x/image/tiff"
)

const (
	topBorder    = 1.0 / 12.0
	bottomBorder = 5.0 / 18.0

	sideBorder = 1.0 / 13.0

	fileSuffix = "pola.tif"
)

var (
	edgeRatio float64
)

func main() {
	flag.Float64Var(&edgeRatio, "r", 1, "edge blur in milli-widths")
	flag.Parse()

	wg := &sync.WaitGroup{}
	jobs := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(jobs, wg)
	}

	for _, fName := range flag.Args() {
		jobs <- fName
	}
	close(jobs)
	wg.Wait()
}

func worker(jobs chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		doOneFrame(job)
	}
}

func doOneFrame(fName string) {
	if strings.HasSuffix(fName, fileSuffix) {
		return
	}

	f, err := os.Open(fName)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	inImage, _, err := image.Decode(f)
	if err != nil {
		log.Println(err)
		return
	}

	inBounds := inImage.Bounds()
	scaledSideBorder := int(float64(inBounds.Dx()) * sideBorder)
	scaledTopBorder := int(float64(inBounds.Dy()) * topBorder)
	scaledBottomBorder := int(float64(inBounds.Dy()) * bottomBorder)

	outBounds := image.Rect(
		inBounds.Min.X-scaledSideBorder, inBounds.Min.Y-scaledTopBorder,
		inBounds.Max.X+scaledSideBorder, inBounds.Max.Y+scaledBottomBorder)

	outImage := image.NewRGBA64(outBounds)

	// Draw background
	white := image.NewUniform(color.White)
	draw.Copy(outImage, outBounds.Min, white, outBounds, draw.Over, nil)

	// Draw main image
	draw.Copy(outImage, inBounds.Min, inImage, inBounds, draw.Over, nil)

	// Create blurred image
	edgePixels := float64(inBounds.Dx()) * (edgeRatio / 1000)
	blur := gaussianBlur(outImage, edgePixels)

	// Create destination mask
	mask := image.NewRGBA64(outBounds)
	draw.Copy(mask, outBounds.Min, image.Opaque, outBounds, draw.Over, nil)
	draw.Copy(mask, inBounds.Min, image.Transparent, inBounds, draw.Src, nil)

	// Overlay blurred image using mask
	draw.Copy(outImage, outBounds.Min, blur, outBounds, draw.Over, &draw.Options{
		DstMask: mask,
	})

	saveImage(outImage, reSuffix(fName, fileSuffix))
}

func saveImage(i image.Image, fName string) {
	w, err := os.Create(fName)
	if err != nil {
		log.Println(err)
		return
	}
	defer w.Close()

	err = tiff.Encode(w, i, &tiff.Options{
		Compression: tiff.Deflate,
		Predictor:   true,
	})
	if err != nil {
		log.Println(err)
	}
}

func reSuffix(fName, suffix string) string {
	segs := strings.Split(fName, ".")
	if len(segs) > 1 {
		segs[len(segs)-1] = suffix
	} else {
		segs = append(segs, suffix)
	}

	return strings.Join(segs, ".")
}
