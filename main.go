package main

import (
	"image/color"
	"log"
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
	cf "github.com/lucasb-eyer/go-colorful"
)

type point struct {
	x, y  int32
	iters int32
}

type bounds struct {
	xmin, xmax float32
	ymin, ymax float32
}

func scaleCoordinate(screenCoord, screenMax, boundMax, boundMin float32) float32 {
	return (screenCoord/screenMax)*(boundMax-(boundMin)) + (boundMin)
}

func mandelBrotIters(x0 float32, y0 float32, maxIter int32) int32 {
	const cutoff = 64
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

func getPoints(screenWidth int32, screenHeight int32, maxIter int32, b bounds, c chan point) {
	var wg sync.WaitGroup
	for x := range screenWidth {
		for y := range screenHeight {
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				scaled_x := scaleCoordinate(float32(x), float32(screenWidth), b.xmax, b.xmin)
				scaled_y := scaleCoordinate(float32(y), float32(screenHeight), b.ymax, b.ymin)
				c <- point{x, y, mandelBrotIters(scaled_x, scaled_y, maxIter)}
			}(&wg)
		}
	}
	wg.Wait()
	close(c)
}

func drawMandelBrot(b bounds, texture rl.Texture2D) {
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()

	data := make([]color.RGBA, screenHeight*screenWidth)
	var maxIter int32
	maxIter = 1000
	c := make(chan point, screenHeight*screenWidth)
	go getPoints(int32(screenWidth), int32(screenHeight), maxIter, b, c)
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
		data[int32(screenWidth)*p.y+p.x] = pointColor
	}

	rl.UpdateTexture(texture, data)
}

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(1920, 1080, "Mandelbrot")
	defer rl.CloseWindow()

	rl.SetTargetFPS(120)

	originalBounds := bounds{-2, 1, -1, 1}
	lastBounds := originalBounds
	currentBounds := originalBounds
	selectedBounds := originalBounds
	boundsChanged := false

	selectionMode := false

	img := rl.LoadImageFromScreen()
	if !rl.IsImageValid(img) {
		log.Fatal("invalid image")
	}
	rl.ImageClearBackground(img, rl.White)
	texture := rl.LoadTextureFromImage(img)
	if !rl.IsTextureValid(texture) {
		log.Fatal("invalid texture")
	}
	rl.UnloadImage(img)

	var mousePos rl.Vector2
	drawMandelBrot(currentBounds, texture)
	for !rl.WindowShouldClose() {

		if rl.IsMouseButtonDown(rl.MouseLeftButton) && !selectionMode {
			selectionMode = true
			mousePos = rl.GetMousePosition()
		}

		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			selectionMode = false
			currentBounds = selectedBounds
		}

		if rl.IsKeyPressed(rl.KeyEqual) {
			currentBounds = originalBounds
		}

		if rl.IsKeyPressed(rl.KeyA) {
			currentBounds = bounds{-5, 5, -1, 1}
		}

		boundsChanged = currentBounds != lastBounds
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		if rl.IsWindowResized() {
			rl.UnloadTexture(texture)
			img := rl.LoadImageFromScreen()
			if !rl.IsImageValid(img) {
				log.Fatal("invalid image")
			}
			rl.ImageClearBackground(img, rl.White)
			texture = rl.LoadTextureFromImage(img)
			if !rl.IsTextureValid(texture) {
				log.Fatal("invalid texture")
			}
			drawMandelBrot(currentBounds, texture)
			rl.UnloadImage(img)
		}
		if boundsChanged {
			// rl.UnloadTexture(m)
			drawMandelBrot(currentBounds, texture)
			lastBounds = currentBounds
			boundsChanged = false
		}
		rl.DrawTexture(texture, 0, 0, rl.White)

		if selectionMode {
			currentPos := rl.GetMousePosition()
			screenSize := rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
			aspect := rl.Vector2Normalize(screenSize)
			diff := rl.Vector2Subtract(currentPos, mousePos)
			length := float32(math.Sqrt(float64(diff.X*diff.X + diff.Y*diff.Y)))
			size := rl.Vector2Scale(aspect, length)
			posX := int32(mousePos.X)
			posY := int32(mousePos.Y)
			width := int32(size.X)
			height := int32(size.Y)
			rl.DrawRectangleLines(posX, posY, width, height, rl.Red)
			selectedBounds.xmin = scaleCoordinate(float32(posX), screenSize.X, currentBounds.xmax, currentBounds.xmin)
			selectedBounds.xmax = scaleCoordinate(float32(posX+width), screenSize.X, currentBounds.xmax, currentBounds.xmin)
			selectedBounds.ymin = scaleCoordinate(float32(posY), screenSize.Y, currentBounds.ymax, currentBounds.ymin)
			selectedBounds.ymax = scaleCoordinate(float32(posY+height), screenSize.Y, currentBounds.ymax, currentBounds.ymin)
		}

		rl.DrawFPS(0, 0)
		rl.EndDrawing()
	}
	rl.UnloadTexture(texture)
}
