package main

import (
	"image/color"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func absSquare(z complex64) float32 {
	return real(z)*real(z) + imag(z)*imag(z)
}

func mandelBrotIters(z complex64) int32 {
	const cutoff = 4
	const maxIter = 50
	c := z
	var i int32
	for i = 0; i < maxIter && absSquare(z) <= cutoff; i++ {
		z = (z * z) + c
	}
	return i
}

type point struct {
	x, y  int32
	iters int32
}

func getPoints(screenWidth int32, screenHeight int32, c chan point) {
	for x := range screenWidth {
		for y := range screenHeight {
			scaled_x := (float32(x)/float32(screenWidth))*(0.47-(-2.0)) + (-2)
			scaled_y := (float32(y)/float32(screenHeight))*(1.12-(-1.12)) + (-1.12)
			var z complex64
			z = complex(scaled_x, scaled_y)
			c <- point{x, y, mandelBrotIters(z)}
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
	c := make(chan point, screenHeight*screenWidth)
	go getPoints(int32(screenWidth), int32(screenHeight), c)
	for p := range c {
		hue := 360.0 * (float32(p.iters) / 50)
		var value float32
		value = 1.0
		if p.iters == 50 {
			value = 0
		}
		pointColor := rl.ColorFromHSV(hue, 1.0, value)
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
		rl.DrawFPS(0, 0)
		if rl.IsWindowResized() {
			rl.UnloadTexture(m)
			m = drawMandelBrot()
		}
		rl.DrawTexture(m, 0, 0, rl.White)
		rl.EndDrawing()
	}
	rl.UnloadTexture(m)
}
