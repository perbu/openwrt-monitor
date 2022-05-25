package matrix

type Pixel struct {
	R, G, B uint8
}

type Matrix struct {
	width, height int
	pixels        []Pixel
}

func Initialize(width, height int) *Matrix {
	m := &Matrix{
		width:  width,
		height: height,
		pixels: make([]Pixel, width*height),
	}
	return m
}

func (m *Matrix) SetPixel(x, y int, p Pixel) {
	m.pixels[y*m.width+x] = p
}

func (m *Matrix) GetPixel(x, y int) Pixel {
	return m.pixels[y*m.width+x]
}

func (m *Matrix) Width() int {
	return m.width
}
func (m *Matrix) Height() int {
	return m.height
}

func (m Matrix) GetMatrix() Matrix {
	return m
}

func (m *Matrix) shiftRight() {
	for y := 0; y < m.height; y++ {
		for x := m.width - 1; x > 0; x-- {
			m.SetPixel(x, y, m.GetPixel(x-1, y))
		}
		m.SetPixel(0, y, Pixel{})
	}
}

func (m *Matrix) PlotNewLine(ingress, egress float64) {
	m.shiftRight()
	ingressPixels := int(ingress * float64(m.height))
	egressPixels := int(egress * float64(m.height))
	for y := 0; y < ingressPixels; y++ {
		intensity := uint8(32 + y*24)
		pixel := m.GetPixel(0, y)
		pixel.R = intensity
		m.SetPixel(0, y, pixel)
	}
	// drawPixels := m.height - egressPixels
	for y := m.height - 1; y >= m.height-egressPixels; y-- {
		intensity := uint8(32 + (m.height-y)*24)
		pixel := m.GetPixel(0, y)
		pixel.B = intensity
		m.SetPixel(0, y, pixel)
	}
}
