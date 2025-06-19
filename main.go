package main

import (
	"image/color"
	"log"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	cf "github.com/lucasb-eyer/go-colorful"
)

func mandelBrotIters(x0 float32, y0 float32, maxIter int32) int32 {
	const cutoff = 4
	var x, y, x2, y2 float32
	var i int32
	x = 0
	y = 0
	x2 = 0
	y2 = 0
	for i = 0; x2+y2 <= cutoff && i < maxIter; i++ {
		y = 2*x*y + y0
		x = x2 - y2 + x0
		x2 = x * x
		y2 = y * y
	}
	return i
}

type point struct {
	x, y  int32
	iters int32
}

func getPoints(screenWidth int32, screenHeight int32, maxIter int32, c chan point) {
	for x := range screenWidth {
		for y := range screenHeight {
			scaled_x := (float32(x)/float32(screenWidth))*(0.47-(-2.0)) + (-2)
			scaled_y := (float32(y)/float32(screenHeight))*(1.12-(-1.12)) + (-1.12)
			c <- point{x, y, mandelBrotIters(scaled_x, scaled_y, maxIter)}
		}
	}
	close(c)
}

func drawMandelBrot() rl.Texture2D {
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()

	img := rl.GenImageColor(screenWidth, screenHeight, color.RGBA{0, 0, 0, 0})
	if !rl.IsImageValid(img) {
		log.Fatalf("Invalid image")
	}
	defer rl.UnloadImage(img)

	var maxIter int32
	maxIter = 1000
	c := make(chan point, screenHeight*screenWidth)
	go getPoints(int32(screenWidth), int32(screenHeight), maxIter, c)
	for p := range c {
		var pointColor color.RGBA
		if p.iters == maxIter {
			pointColor = rl.Black
		} else {
			i := float64(p.iters)
			maxI := float64(maxIter)
			s := i / maxI
			v := 1.0 - (math.Cos(math.Pi*s) * math.Cos(math.Pi*s))
			L := 75 - (75 * v)
			C := 28 + (75 - (75 * v))
			H := math.Mod(math.Pow(360*s, 1.5), 360)
			lch := cf.LuvLCh(L, C, H)
			r, g, b, a := lch.RGBA()
			pointColor.R = uint8(r)
			pointColor.G = uint8(g)
			pointColor.B = uint8(b)
			pointColor.A = uint8(a)
		}
		rl.ImageDrawPixel(img, p.x, p.y, pointColor)
	}

	texture := rl.LoadTextureFromImage(img)
	if !rl.IsTextureValid(texture) {
		log.Fatalf("Invalid texture")
	}
	return texture
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(1920, 1080, "Mandelbrot")
	defer rl.CloseWindow()

	rl.SetTargetFPS(120)

	m := drawMandelBrot()
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		if rl.IsWindowResized() {
			rl.UnloadTexture(m)
			m = drawMandelBrot()
		}
		rl.DrawTexture(m, 0, 0, rl.White)
		rl.DrawFPS(0, 0)
		rl.EndDrawing()
	}
	rl.UnloadTexture(m)
}
