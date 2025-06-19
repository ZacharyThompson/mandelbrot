package main

import (
	"image/color"
	"log"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//	func mandelBrotIters(z complex128) int {
//		const cutoff = 400
//		const max_iter = 1000
//		const c = 0 + 0i
//		var i int
//		for i = 0; i < iterations && cmplx.Abs(z) <= cutoff; i++ {
//			z = (z * z) + c
//		}
//		return cmplx.Abs(z) <= cutoff
//	}
func absSquare(z complex128) float64 {
	return math.Pow(real(z), 2) + math.Pow(imag(z), 2)
}

func inMandelbrotSet(z complex128) bool {
	const cutoff = 4
	const iterations = 100
	c := z
	for i := 0; i < iterations && absSquare(z) <= cutoff; i++ {
		z = (z * z) + c
	}
	return absSquare(z) <= cutoff
}

type point struct {
	x, y int
}

func getPoints(screenWidth int, screenHeight int, c chan point) {
	for x := range screenWidth {
		for y := range screenHeight {
			scaled_x := (float64(x)/float64(screenWidth))*(0.47-(-2.0)) + (-2)
			scaled_y := (float64(y)/float64(screenHeight))*(1.12-(-1.12)) + (-1.12)
			z := complex(scaled_x, scaled_y)
			if inMandelbrotSet(z) {
				c <- point{x, y}
			}
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
	go getPoints(screenWidth, screenHeight, c)
	for p := range c {
		rl.ImageDrawPixel(img, int32(p.x), int32(p.y), rl.Black)
	}

	texture := rl.LoadTextureFromImage(img)
	if !rl.IsTextureValid(texture) {
		log.Fatalf("Invalid texture")
	}
	return texture
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(1920, 1080, "mandlebrot")
	defer rl.CloseWindow()

	rl.SetTargetFPS(120)

	m := drawMandelBrot()
	for !rl.WindowShouldClose() {
		// screenSize := rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
		// origin := rl.Vector2Scale(screenSize, 0.5)
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
