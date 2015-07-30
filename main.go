// for_test project main.go
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"

	"strconv"
	"strings"
)

const (
	power = 9 //good result if eq 8

	rougness   = 0.9
	baseHeight = 255
)

var (
	newImage *image.RGBA
	white    color.Color = color.RGBA{255, 255, 255, 255}
	black    color.Color = color.RGBA{0, 0, 0, 255}
	blue     color.Color = color.RGBA{0, 0, 255, 255}

	hmap       [][]int
	height     = int(math.Pow(2, float64(power))) + 1 //513 257
	width      = height
	smotheStep = int(math.Pow(2, float64(power-5)))
)

func calcRandom(l int) int {
	intn := (float32(l*2) * rougness / float32(height)) * baseHeight * 2

	random := 0
	if int(intn) > 0 {
		random = (rand.Intn(int(intn)) - int(intn/2))
	}
	return random
}

func square(lx, ly, rx, ry int) {
	l := (rx - lx) / 2

	a := hmap[lx][ly]
	b := hmap[lx][ry]
	c := hmap[rx][ly]
	d := hmap[rx][ry]

	cx := lx + l
	cy := ly + l

	hmap[cx][cy] = (a+b+c+d)/4 + calcRandom(l)
}

func diamond(x, y, l int) {
	lx := x - l
	rx := x + l
	ty := y - l
	by := y + l

	a := 0
	if lx > 0 {
		a = hmap[lx][y]
	}
	b := 0
	if ty > 0 {
		b = hmap[x][ty]
	}
	c := 0
	if rx < width-1 {
		c = hmap[rx][y]
	}
	d := 0
	if by < height-1 {
		d = hmap[x][by]
	}

	hmap[x][y] = (a+b+c+d)/4 + +calcRandom(l)
}

func diamondSquare(lx, ly, rx, ry int) {
	l := (rx - lx) / 2

	square(lx, ly, rx, ry)

	diamond(lx, ly+l, l)
	diamond(lx, ry-l, l)
	diamond(lx+l, ry, l)
	diamond(rx-l, ry, l)
}

func generate() {
	//Set predefault
	hmap[0][0] = 0              //rand.Intn(64)
	hmap[0][width-1] = 0        //rand.Intn(64)
	hmap[height-1][0] = 0       //rand.Intn(64)
	hmap[height-1][width-1] = 0 //rand.Intn(baseHeight)

	for step := (width - 1) / 2; step > 0; step /= 2 {
		for y := 0; y < height-1; y += step {
			for x := 0; x < width-1; x += step {
				diamondSquare(x, y, x+step, y+step)
			}
		}
	}
}

func smoothMap() {
	log.Println("smoothMap")
	for step := smotheStep; step > 0; step /= 2 {
		for y := 0; y < height-1; y += step {
			for x := 0; x < width-1; x += step {
				diamondSquare(x, y, x+step, y+step)
			}
		}
	}
}

func makeHMap() {
	log.Println("makeHMap")
	hmap = make([][]int, height)
	for i := 0; i < height; i++ {
		hmap[i] = make([]int, width)
	}
}

func showHMap() {
	for i := range hmap {
		fmt.Print("[")
		for j := range hmap[i] {
			fmt.Print(hmap[i][j], " ")
		}
		fmt.Println("]")
	}
}

func normaliseHMap() {
	log.Println("normaliseHMap")
	max, min := func() (int, int) {
		max, min := 0, 0
		for i := range hmap {
			for j := range hmap[i] {
				if max < hmap[i][j] {
					max = hmap[i][j]
				}
				if min > hmap[i][j] {
					min = hmap[i][j]
				}
			}
		}
		return max, min
	}()

	delta := max - min
	percent := float32(baseHeight) / float32(delta)

	for i := range hmap {
		for j := range hmap[i] {
			hmap[i][j] = int(float32(hmap[i][j]-min) * percent)
		}
	}
}

func saveImage(name string) {
	var waterLevel = (hmap[0][0]+hmap[0][width-1]+hmap[height-1][0]+hmap[height-1][width-1])/4 + 1

	log.Println("Save image: ", name)

	newImage = image.NewRGBA(image.Rect(0, 0, height, width))
	draw.Draw(newImage, newImage.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)

	for i := range hmap {
		for j := range hmap[i] {
			//			newImage.Set(i, j, color.RGBA{uint8(hmap[i][j]), 0, 0, 255})
			switch {
			case hmap[i][j] < waterLevel:
				newImage.Set(i, j, color.RGBA{0, 0, uint8(hmap[i][j]), 255})
			case hmap[i][j] >= waterLevel && hmap[i][j] < 170:
				newImage.Set(i, j, color.RGBA{0, 255 - uint8(hmap[i][j]), 0, 255})
			case hmap[i][j] >= 170 && hmap[i][j] < 220:
				newImage.Set(i, j, color.RGBA{255 - uint8(hmap[i][j]), 220 - uint8(hmap[i][j]), 0, 255})
			case hmap[i][j] >= 220:
				newImage.Set(i, j, color.RGBA{uint8(hmap[i][j]), uint8(hmap[i][j]), uint8(hmap[i][j]), 255})
			}
		}
	}
	w, _ := os.Create(name)
	defer w.Close()
	png.Encode(w, newImage) //Encode writes the Image m to w in PNG format.
}

func generateNMaps(seed, n int) {
	if n <= 0 {
		log.Panicln("n should be great then 0")
	}
	log.Println("Start generation")

	for j := 0; j < n; j++ {
		rand.Seed(int64(seed + j))

		makeHMap()
		var generateSteps = 32
		for i := 0; i < generateSteps; i++ {
			log.Println("Generate progress - total:", generateSteps, " step: ", i)
			generate()
		}
		smoothMap()
		normaliseHMap()
		//	showHMap()

		saveImage(strings.Join([]string{"name", strconv.Itoa(j), ".png"}, ""))
	}
}

func main() {
	generateNMaps(42, 20)
}
