package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nathany/bobblehat/sense/screen"
	"github.com/nathany/bobblehat/sense/screen/color"
	"github.com/perbu/go-matrix/matrix"
	"github.com/perbu/go-matrix/router"
	"log"
	"time"
)

func mainLoop(r *router.Router) {
	m := matrix.Initialize(8, 8)
	fb := screen.NewFrameBuffer()
	for {
		stats := r.GetTrafficStats()
		fmt.Printf("tx: %010d / %010d  rx: %010d/ %010d [duration: %v]\n", stats.CurTx, stats.MaxTx, stats.CurRx, stats.MaxRx, stats.Duration)
		m.PlotNewLine(float64(stats.CurTx)/float64(stats.MaxTx), float64(stats.CurRx)/float64(stats.MaxRx))
		cur := m.GetMatrix()
		err := piRender(fb, cur)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second)
	}
}

func piRender(fb *screen.FrameBuffer, cur matrix.Matrix) error {
	err := screen.Clear()
	if err != nil {
		return fmt.Errorf("error clearing screen: %w", err)
	}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			pixel := cur.GetPixel(x, y)
			fb.SetPixel(y, x, color.New(pixel.R, pixel.G, pixel.B))
		}
	}
	err = screen.Draw(fb)
	if err != nil {
		return fmt.Errorf("error drawing framebuffer: %w", err)
	}
	return nil
}

func run() {
	r := router.New("http://10.0.0.1/")
	_ = r.GetTrafficStats()
	mainLoop(r)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Could not load .env file")
	}
	run()
}
