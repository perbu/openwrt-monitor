package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nathany/bobblehat/sense/screen"
	"github.com/nathany/bobblehat/sense/screen/color"
	"github.com/perbu/go-matrix/matrix"
	"github.com/perbu/go-matrix/router"
	"github.com/perbu/go-matrix/tui"
	"log"
	"os"
	"os/signal"
	"time"
)

func mainLoop(r *router.Router) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	m := matrix.Initialize(8, 8)
	fb := screen.NewFrameBuffer()
	for ctx.Err() == nil {
		stats := r.GetTrafficStats()
		fmt.Printf("tx: %010d / %010d  rx: %010d/ %010d [duration: %v]\n", stats.CurTx, stats.MaxTx, stats.CurRx, stats.MaxRx, stats.Duration)
		m.PlotNewLine(float64(stats.CurTx)/float64(stats.MaxTx), float64(stats.CurRx)/float64(stats.MaxRx))
		cur := m.GetMatrix()
		err := piRender(fb, cur)
		if err != nil {
			return fmt.Errorf("piRender: %w", err)
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil
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

func verifyEnv() error {
	required := []string{"ROUTER_USER", "ROUTER_PASSWORD", "ROUTER_URL"}
	for _, env := range required {
		if _, ok := os.LookupEnv(env); !ok {
			return fmt.Errorf("environment variable %s is required", env)
		}
	}
	return nil
}

func run() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("No env file found. Relying on environment variables")
	}
	err = verifyEnv()
	if err != nil {
		return fmt.Errorf("error verifying environment: %w", err)
	}
	r := router.New(os.Getenv("ROUTER_URL"))
	_ = r.GetTrafficStats()
	if len(os.Args) < 2 {
		return fmt.Errorf("missing argument ('hat' or 'tui')")
	}
	switch os.Args[1] {
	case "tui":
		return tui.Run(r)
	case "hat":
		return mainLoop(r)
	default:
		return fmt.Errorf("wrong argument (must be 'hat' or 'tui')")
	}
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
