package main

import (
	"image/color"
	"log"
	"math"
	"sync"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	cf "github.com/lucasb-eyer/go-colorful"
)

type imgPoint struct {
	x, y int
}

type Plotter struct {
	colorMap               map[int]color.RGBA
	data                   map[imgPoint]int
	xMin, xMax, yMin, yMax float64
	imgWidth, imgHeight    int
}

func (p *Plotter) computeData() {
	maxIter := 1000
	var wg sync.WaitGroup
	nRoutines := p.imgHeight / 3
	rowsPerRoutine := p.imgHeight / nRoutines
	for i := range nRoutines {
		start := i * rowsPerRoutine
		end := start + rowsPerRoutine
		if i == nRoutines-1 {
			// last goroutine calculates remainder
			end = p.imgHeight
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := range p.imgWidth {
					scaled_x := scaleCoordinate(float64(x), float64(p.imgWidth), p.xMax, p.xMin)
					scaled_y := scaleCoordinate(float64(y), float64(p.imgHeight), p.yMax, p.yMin)
					p.data[imgPoint{x, y}] = p.mandelBrotIters(scaled_x, scaled_y, maxIter)
				}
			}
		}(&wg)
	}
	wg.Wait()
}

func (p *Plotter) getColor(iters, maxIter int) color.RGBA {
	c, ok := p.colorMap[iters]
	if !ok {
		if iters == maxIter {
			c = color.RGBA{0, 0, 0, 0xff}
		} else {
			i := float64(iters)
			maxI := float64(maxIter)
			s := i / maxI
			v := 1.0 - (math.Cos(math.Pi*s) * math.Cos(math.Pi*s))
			L := 75 - (75 * v)
			C := 28 + (75 - (75 * v))
			H := math.Mod(math.Pow(360*s, 1.5), 360)
			lch := cf.LuvLCh(L, C, H)
			r, g, b, a := lch.RGBA()
			c.R = uint8(r)
			c.G = uint8(g)
			c.B = uint8(b)
			c.A = uint8(a)
		}
		p.colorMap[iters] = c
	}
	return c
}

func (p *Plotter) mandelBrotIters(x0 float64, y0 float64, maxIter int) int {
	const cutoff = 4
	var x, y, x2, y2 float64
	x = 0
	y = 0
	x2 = 0
	y2 = 0
	var i int
	for i = 0; x2+y2 <= cutoff && i < maxIter; i++ {
		y = 2*x*y + y0
		x = x2 - y2 + x0
		x2 = x * x
		y2 = y * y
	}
	return i
}

func (p *Plotter) drawMandelBrot(enableColor bool, b Bounds) rl.Image {
	start := time.Now()
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()

	maxIter := 2000
	data := p.computeData(screenWidth, screenHeight, maxIter, b)
	colors := make([]color.RGBA, screenWidth*screenHeight)
	for y := range screenHeight {
		for x := range screenWidth {
			iters := data[y*screenWidth+x]
			var c color.RGBA
			if !enableColor && iters != maxIter {
				c = color.RGBA{0xff, 0xff, 0xff, 0xff}
			} else {
				c = getColor(iters, maxIter)
			}
			colors[y*screenWidth+x] = c
		}
	}
	rl.UpdateTexture(texture, colors)
	log.Println("Time to render: ", time.Since(start))
}
