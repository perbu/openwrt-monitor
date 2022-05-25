package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/perbu/go-matrix/matrix"
	"github.com/perbu/go-matrix/router"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	r := router.New("http://10.0.0.1")
	_ = r.GetTrafficStats() // get initial counters so the next call will be accurate
	p := tea.NewProgram(initialModel(r))
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

type model struct {
	matrix *matrix.Matrix
	debug  string
	router *router.Router
}

type tickMsg time.Time

func (m *model) tickCmd() tea.Cmd {

	return tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel(r *router.Router) model {
	return model{
		matrix: matrix.Initialize(64, 16),
		router: r,
	}
}

func (m model) Init() tea.Cmd {
	return m.tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tickMsg:
		stats := m.router.GetTrafficStats()
		m.debug = fmt.Sprintf("tx: %010d / %010d  rx: %010d/ %010d [duration: %v]", stats.CurTx, stats.MaxTx, stats.CurRx, stats.MaxRx, stats.Duration)
		m.matrix.PlotNewLine(float64(stats.CurTx)/float64(stats.MaxTx), float64(stats.CurRx)/float64(stats.MaxRx))
		return m, m.tickCmd()
	}

	// Just return `nil`, which means "no I/O right now, please."
	return m, nil
}

func (m model) View() string {
	var s string
	s += m.debug + "\n"
	s += RenderMatrix(m.matrix)
	return s
}

func RenderPixel(p matrix.Pixel) string {
	color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", p.R, p.G, p.B))
	style := lipgloss.NewStyle().Background(color)
	return style.Render("  ") // more or less a square.
}

func RenderMatrix(m *matrix.Matrix) string {
	var s string
	for y := 0; y < m.Height(); y++ {
		for x := 0; x < m.Width(); x++ {
			s += RenderPixel(m.GetPixel(x, y))
		}
		s += "\n"
	}
	return s
}
