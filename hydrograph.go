package main

// hydrograph.go — firma visual del informe: el día dibujado como un
// hidrógrafo de estación de aforo. Curva de ingesta acumulada contra la
// línea de capacidad (meta), sellos de clima, numeración monoespaciada.
// Disciplinas donadas: quietud de papel extensa alrededor de una sola curva
// (saville-catalog), densidad numérica mono con filetes a 1 px
// (ikeda-datamatics), el día como recipiente de capacidad fija con marca de
// sobrellenado (cassette-j-card).

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

// WaterLog es un evento de ingesta intradía (RFC3339 + ml).
type WaterLog struct {
	At string `json:"at"`
	Ml int    `json:"ml"`
}

// ---------- helpers de pixel (compartidos con el resumen semanal) ----------

func repFillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && y >= 0 && x < img.Bounds().Dx() && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

func repHLine(img *image.RGBA, x0, x1, y int, c color.Color) {
	for x := x0; x < x1; x++ {
		img.Set(x, y, c)
	}
}

func repVLine(img *image.RGBA, x, y0, y1 int, c color.Color) {
	for y := y0; y < y1; y++ {
		img.Set(x, y, c)
	}
}

func repDashedHLine(img *image.RGBA, x0, x1, y, on, off int, c color.Color) {
	x := x0
	for x < x1 {
		end := x + on
		if end > x1 {
			end = x1
		}
		repHLine(img, x, end, y, c)
		x += on + off
	}
}

func repRectStroke(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	repHLine(img, x0, x1+1, y0, c)
	repHLine(img, x0, x1+1, y1, c)
	repVLine(img, x0, y0, y1+1, c)
	repVLine(img, x1, y0, y1+1, c)
}

// ---------- widget ----------

// Hydrograph dibuja el día como curva de aforo: ingesta acumulada por hora
// contra la meta (línea de capacidad), con sello de clima cuando hay bonus.
type Hydrograph struct {
	widget.BaseWidget
	logs     []WaterLog
	consumed int
	goal     int
	bonus    int

	rast   *canvas.Raster
	read   *canvas.Text
	stamp  *canvas.Text
	stampB *canvas.Rectangle
	trace  *canvas.Text
}

func NewHydrograph() *Hydrograph {
	g := &Hydrograph{}
	g.ExtendBaseWidget(g)

	g.rast = canvas.NewRaster(g.draw)

	g.read = canvas.NewText("0 / 0 ml", repInk)
	g.read.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	g.read.TextSize = 15
	g.read.Alignment = fyne.TextAlignLeading

	g.stamp = canvas.NewText("", repVermilion)
	g.stamp.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	g.stamp.TextSize = 10
	g.stamp.Alignment = fyne.TextAlignTrailing
	g.stamp.Hidden = true

	g.stampB = canvas.NewRectangle(color.Transparent)
	g.stampB.StrokeColor = repVermilion
	g.stampB.StrokeWidth = 1
	g.stampB.Hidden = true

	g.trace = canvas.NewText("", repInkMuted)
	g.trace.TextStyle = fyne.TextStyle{Monospace: true}
	g.trace.TextSize = 9
	g.trace.Alignment = fyne.TextAlignLeading
	g.trace.Hidden = true

	return g
}

// SetDay alimenta el hidrógrafo con datos reales del día.
func (g *Hydrograph) SetDay(logs []WaterLog, consumed, goal, bonus int) {
	g.logs = logs
	g.consumed = consumed
	g.goal = goal
	g.bonus = bonus

	pct := 0.0
	if goal > 0 {
		pct = float64(consumed) / float64(goal) * 100
	}
	g.read.Text = fmt.Sprintf("%d / %d ml · %.0f%%", consumed, goal, pct)

	if bonus > 0 {
		g.stamp.Text = fmt.Sprintf("%s +%d", T("climate"), bonus)
		g.stamp.Hidden = false
		g.stampB.Hidden = false
	} else {
		g.stamp.Text = ""
		g.stamp.Hidden = true
		g.stampB.Hidden = true
	}
	if consumed > 0 && len(logs) == 0 {
		g.trace.Text = T("no_traces")
		g.trace.Hidden = false
	} else {
		g.trace.Hidden = true
	}

	canvas.Refresh(g.rast)
	canvas.Refresh(g.read)
	canvas.Refresh(g.stamp)
	canvas.Refresh(g.stampB)
	canvas.Refresh(g.trace)
}

func (g *Hydrograph) CreateRenderer() fyne.WidgetRenderer {
	return &hydroRender{g: g, objs: []fyne.CanvasObject{
		g.rast, g.stampB, g.stamp, g.trace, g.read,
	}}
}

type hydroRender struct {
	g    *Hydrograph
	objs []fyne.CanvasObject
}

func (r *hydroRender) Layout(s fyne.Size) {
	r.g.rast.Resize(s)
	r.g.rast.Move(fyne.NewPos(0, 0))

	const padX, readY, readH = 12.0, 10.0, 22.0
	r.g.read.Resize(fyne.NewSize(s.Width-padX*2, readH))
	r.g.read.Move(fyne.NewPos(padX, readY))

	if !r.g.stamp.Hidden {
		stampW := float32(len(r.g.stamp.Text))*6.4 + 14
		r.g.stamp.Resize(fyne.NewSize(stampW-10, 16))
		r.g.stamp.Move(fyne.NewPos(s.Width-padX-stampW+5, readY+2))
		r.g.stampB.Resize(fyne.NewSize(stampW, 20))
		r.g.stampB.Move(fyne.NewPos(s.Width-padX-stampW, readY))
	}

	r.g.trace.Resize(fyne.NewSize(s.Width-padX*2, 14))
	r.g.trace.Move(fyne.NewPos(padX+4, s.Height-18))
}

func (r *hydroRender) MinSize() fyne.Size { return fyne.NewSize(320, 180) }

func (r *hydroRender) Refresh() {
	canvas.Refresh(r.g.rast)
	canvas.Refresh(r.g.read)
	canvas.Refresh(r.g.stamp)
	canvas.Refresh(r.g.stampB)
	canvas.Refresh(r.g.trace)
}

func (r *hydroRender) Objects() []fyne.CanvasObject { return r.objs }
func (r *hydroRender) Destroy()                     {}

// draw rasteriza el panel: papel de panel, retícula de reglas, línea de
// capacidad punteada y la curva escalonada del día.
func (g *Hydrograph) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	repFillRect(img, 0, 0, w, h, repPanel)
	repRectStroke(img, 0, 0, w-1, h-1, repRule)

	const left, right, top, bottom = 12, 12, 40, 18
	plotW := w - left - right
	plotH := h - top - bottom
	if plotW < 40 || plotH < 24 {
		return img
	}

	// Retícula: reglas horizontales a 1/4, 2/4, 3/4 y verticales a 6/12/18 h.
	for i := 1; i <= 3; i++ {
		y := top + plotH*i/4
		repHLine(img, left, left+plotW, y, repHairline)
		x := left + plotW*i/4
		repVLine(img, x, top, top+plotH, repHairline)
	}
	// Línea base (0 ml).
	repHLine(img, left, left+plotW, top+plotH, repRule)

	// Escala: capacidad = meta, redondeada a múltiplos de 500.
	maxVal := g.goal
	if g.consumed > maxVal {
		maxVal = g.consumed
	}
	if maxVal <= 0 {
		maxVal = 2000
	}
	maxVal = ((maxVal + 499) / 500) * 500

	toY := func(ml int) int {
		return top + plotH - int(float64(ml)/float64(maxVal)*float64(plotH))
	}

	// Línea de capacidad (meta) punteada en vermellón.
	goalY := 0
	if g.goal > 0 {
		goalY = toY(g.goal)
		if goalY >= top && goalY <= top+plotH {
			repDashedHLine(img, left, left+plotW, goalY, 6, 5, repVermilion)
		}
	}

	// Puntos de la curva: (minuto del día, ml acumulado).
	type pt struct{ x, y int }
	nowMin := float64(time.Now().Hour()*60+time.Now().Minute()) / 1440.0
	if nowMin <= 0 {
		nowMin = 1.0 / 1440.0
	}
	toX := func(min float64) int {
		return left + int(min*float64(plotW))
	}

	pts := make([]pt, 0, len(g.logs)+2)
	pts = append(pts, pt{left, top + plotH})
	cum := 0
	curMin := 0.0
	for _, l := range g.logs {
		t, err := time.Parse(time.RFC3339, l.At)
		if err != nil {
			// Tolerante a sellos sin offset (datos legados de la semilla).
			t, err = time.Parse("2006-01-02T15:04:05", l.At)
		}
		if err != nil || l.Ml <= 0 {
			continue
		}
		m := float64(t.Hour()*60+t.Minute()) / 1440.0
		if m < curMin {
			m = curMin
		}
		cum += l.Ml
		x := toX(m)
		if x > left+plotW {
			x = left + plotW
		}
		pts = append(pts, pt{x, toY(cum)})
		curMin = m
	}
	endX := toX(nowMin)
	if endX > left+plotW {
		endX = left + plotW
	}
	if endX < left {
		endX = left
	}
	lastY := pts[len(pts)-1].y
	pts = append(pts, pt{endX, lastY})

	// Área bajo la curva (lavado hidro) columna por columna.
	for i := 1; i < len(pts); i++ {
		x0, x1 := pts[i-1].x, pts[i].x
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		for x := x0; x < x1 && x < left+plotW; x++ {
			if x < left {
				continue
			}
			for y := pts[i-1].y + 1; y < top+plotH; y++ {
				img.Set(x, y, repHydroSoft)
			}
		}
	}

	// Curva escalonada: tramo horizontal en el nivel previo, salto vertical.
	stroke := func(x0, x1, y int) {
		if y < top {
			y = top
		}
		for x := x0; x < x1; x++ {
			img.Set(x, y, repHydro)
			img.Set(x, y+1, repHydro)
		}
	}
	rise := func(x, y0, y1 int) {
		if y0 > y1 {
			y0, y1 = y1, y0
		}
		for y := y0; y <= y1; y++ {
			if y < top {
				continue
			}
			img.Set(x, y, repHydro)
			img.Set(x+1, y, repHydro)
		}
	}
	for i := 1; i < len(pts); i++ {
		stroke(pts[i-1].x, pts[i].x, pts[i-1].y)
		if pts[i].y != pts[i-1].y {
			rise(pts[i].x, pts[i-1].y, pts[i].y)
		}
	}

	// Marca de sobrellenado: tramos de curva por encima de la capacidad.
	if g.goal > 0 && goalY >= top {
		for i := 1; i < len(pts); i++ {
			y0, y1 := pts[i-1].y, pts[i].y
			if y0 > y1 {
				y0, y1 = y1, y0
			}
			if y0 < goalY {
				yEnd := y1
				if yEnd > goalY {
					yEnd = goalY
				}
				for y := y0; y <= yEnd; y++ {
					if y < top {
						continue
					}
					x := pts[i].x
					if pts[i].x > pts[i-1].x {
						x = pts[i].x - 1
					}
					img.Set(x, y, repVermilion)
					img.Set(x-1, y, repVermilion)
				}
			}
		}
	}

	// Etiquetas de escala (mono, tinta desvanecida): máximo y 0.
	// Se dejan al texto del widget; aquí solo la retícula (raster sin texto).

	return img
}
