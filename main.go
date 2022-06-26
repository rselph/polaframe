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

func main() {
	flag.Parse()

	wg := &sync.WaitGroup{}
	jobs := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(jobs, wg)
	}

	for _, fname := range flag.Args() {
		jobs <- fname
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

func doOneFrame(fname string) {
	if strings.HasSuffix(fname, fileSuffix) {
		return
	}

	f, err := os.Open(fname)
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

	background := image.NewUniform(color.White)

	outImage := image.NewRGBA64(outBounds)
	draw.Copy(outImage, outBounds.Min, background, outBounds, draw.Over, nil)
	draw.Copy(outImage, inBounds.Min, inImage, inBounds, draw.Over, nil)

	fname = reSuffix(fname, fileSuffix)
	w, err := os.Create(fname)
	if err != nil {
		log.Println(err)
		return
	}
	defer w.Close()

	err = tiff.Encode(w, outImage, &tiff.Options{
		Compression: tiff.Deflate,
		Predictor:   true,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func reSuffix(fname, suffix string) string {
	segs := strings.Split(fname, ".")
	if len(segs) > 1 {
		segs[len(segs)-1] = suffix
	} else {
		segs = append(segs, suffix)
	}

	return strings.Join(segs, ".")
}
