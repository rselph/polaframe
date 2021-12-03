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
	thinBorder  = 0.05
	thickBorder = 0.277778
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
	var scaledThinBorder int
	if inBounds.Dx() < inBounds.Dy() {
		scaledThinBorder = int(float64(inBounds.Dx()) * thinBorder)
	} else {
		scaledThinBorder = int(float64(inBounds.Dy()) * thinBorder)
	}
	scaledThickBorder := int(float64(inBounds.Dy()) * thickBorder)

	outBounds := image.Rect(
		inBounds.Min.X-scaledThinBorder, inBounds.Min.Y-scaledThinBorder,
		inBounds.Max.X+scaledThinBorder, inBounds.Max.Y+scaledThickBorder)

	background := image.NewUniform(color.White)

	outImage := image.NewRGBA64(outBounds)
	draw.Copy(outImage, outBounds.Min, background, outBounds, draw.Over, nil)
	draw.Copy(outImage, inBounds.Min, inImage, inBounds, draw.Over, nil)

	fname = reSuffix(fname, "pola.tiff")
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
