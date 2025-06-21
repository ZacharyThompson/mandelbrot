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

type point struct {
	x, y  int32
	iters int32
}

type Bounds struct {
	xmin, xmax float64
	ymin, ymax float64
}

func scaleCoordinate(screenCoord, screenMax, boundMax, boundMin float64) float64 {
	return (screenCoord/screenMax)*(boundMax-(boundMin)) + (boundMin)
}

var colorMap map[int]color.RGBA

func getColor(iters, maxIter int) color.RGBA {
	c, ok := colorMap[iters]
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
		colorMap[iters] = c
	}
	return c
}

func mandelBrotIters(x0 float64, y0 float64, maxIter int) int {
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

func getPoints(width int, height int, maxIter int, b Bounds) []int {
	data := make([]int, width*height)
	var wg sync.WaitGroup
	nRoutines := height / 3
	rowsPerRoutine := height / nRoutines
	for i := range nRoutines {
		start := i * rowsPerRoutine
		end := start + rowsPerRoutine
		if i == nRoutines-1 {
			// last goroutine calculates remainder
			end = height
		}
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			for y := start; y < end; y++ {
				for x := range width {
					scaled_x := scaleCoordinate(float64(x), float64(width), b.xmax, b.xmin)
					scaled_y := scaleCoordinate(float64(y), float64(height), b.ymax, b.ymin)
					data[y*width+x] = mandelBrotIters(scaled_x, scaled_y, maxIter)
				}
			}
		}(&wg)
	}
	wg.Wait()
	return data
}

func drawMandelBrot(enableColor bool, b Bounds, texture rl.Texture2D) {
	start := time.Now()
	screenWidth := rl.GetScreenWidth()
	screenHeight := rl.GetScreenHeight()

	maxIter := 2000
	data := getPoints(screenWidth, screenHeight, maxIter, b)
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

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(1080, 1080, "Mandelbrot")
	defer rl.CloseWindow()

	rl.SetTargetFPS(120)

	colorMap = map[int]color.RGBA{}
	originalBounds := Bounds{-2.2, 2.2, -2.2, 2.2}
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

	enableColor := true
	var mousePos rl.Vector2
	drawMandelBrot(enableColor, currentBounds, texture)
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
			currentBounds = Bounds{-5, 5, -1, 1}
		}

		if rl.IsKeyPressed(rl.KeyC) {
			enableColor = !enableColor
			drawMandelBrot(enableColor, currentBounds, texture)
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
			drawMandelBrot(enableColor, currentBounds, texture)
			rl.UnloadImage(img)
		}
		if boundsChanged {
			drawMandelBrot(enableColor, currentBounds, texture)
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
			rl.DrawRectangle(posX, posY, width, height, color.RGBA{0x00, 0xaa, 0xff, 0x88})
			selectedBounds.xmin = scaleCoordinate(float64(posX), float64(screenSize.X), currentBounds.xmax, currentBounds.xmin)
			selectedBounds.xmax = scaleCoordinate(float64(posX+width), float64(screenSize.X), currentBounds.xmax, currentBounds.xmin)
			selectedBounds.ymin = scaleCoordinate(float64(posY), float64(screenSize.Y), currentBounds.ymax, currentBounds.ymin)
			selectedBounds.ymax = scaleCoordinate(float64(posY+height), float64(screenSize.Y), currentBounds.ymax, currentBounds.ymin)
		}

		rl.DrawFPS(0, 0)
		rl.EndDrawing()
	}
	rl.UnloadTexture(texture)
}
