package main

import (
	"image/color"
	"log"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Bounds struct {
	xmin, xmax float64
	ymin, ymax float64
}

func scaleCoordinate(screenCoord, screenMax, boundMax, boundMin float64) float64 {
	return (screenCoord/screenMax)*(boundMax-(boundMin)) + (boundMin)
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
