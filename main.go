package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"strings"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Mood int

const (
	MoodWell    Mood = iota
	MoodNeutral
	MoodLow
	MoodTense
)

func (m Mood) String() string {
	switch m {
	case MoodWell:
		return "Bien"
	case MoodNeutral:
		return "Neutral"
	case MoodLow:
		return "Bajo"
	case MoodTense:
		return "Tenso"
	}
	return "Neutral"
}

type Settings struct {
	Weight           int     `json:"weight"`
	Height           int     `json:"height"`
	TodayConsumed    int     `json:"today_consumed"`
	LastResetDate    string  `json:"last_reset_date"`
	Language         string  `json:"language"`
	LocationName     string  `json:"location_name"`
	LocationLat      float64 `json:"location_lat"`
	LocationLon      float64 `json:"location_lon"`
	ReminderInterval int     `json:"reminder_interval"`
	CurrentMood      Mood    `json:"current_mood"`
}

type HistoryEntry struct {
	Date string `json:"date"`
	Ml   int    `json:"ml"`
	Mood string `json:"mood,omitempty"`
}

type WeatherData struct {
	Temperature float64
	Humidity    float64
}

type GeoResult struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type GeoResponse struct {
	Results []GeoResult `json:"results"`
}

type WeatherResponse struct {
	Current struct {
		Temperature float64 `json:"temperature_2m"`
		Humidity    float64 `json:"relative_humidity_2m"`
	} `json:"current"`
}

type HydrapotionApp struct {
	settings    Settings
	history     []HistoryEntry
	mu          sync.RWMutex
	dataDir     string
	reminder    *time.Timer
	onUpdate    func()
	mainWindow  fyne.Window
	settingsWin fyne.Window
	reminderWin fyne.Window
	app         fyne.App
	weather     *WeatherData
}

func NewHydrapotionApp(a fyne.App) *HydrapotionApp {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".hydrapotion")
	os.MkdirAll(dataDir, 0755)

	h := &HydrapotionApp{
		dataDir: dataDir,
		app:     a,
		settings: Settings{
			Weight:           70,
			Height:           170,
			Language:         "es",
			ReminderInterval: 1800,
			CurrentMood:      MoodNeutral,
		},
	}

	h.loadSettings()
	h.loadHistory()
	h.resetIfNewDay()

	for _, e := range h.history {
		if e.Date == h.todayStr() && e.Mood != "" {
			for i, m := range []string{"Bien", "Neutral", "Bajo", "Tenso"} {
				if e.Mood == m {
					h.settings.CurrentMood = Mood(i)
					break
				}
			}
			break
		}
	}

	if h.settings.LocationName != "" {
		go h.fetchWeather()
	}

	return h
}

func (h *HydrapotionApp) dataFile(name string) string {
	return filepath.Join(h.dataDir, name)
}

func (h *HydrapotionApp) loadSettings() {
	data, err := os.ReadFile(h.dataFile("settings.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, &h.settings)
}

func (h *HydrapotionApp) saveSettings() {
	data, _ := json.MarshalIndent(h.settings, "", " ")
	os.WriteFile(h.dataFile("settings.json"), data, 0644)
}

func (h *HydrapotionApp) loadHistory() {
	data, err := os.ReadFile(h.dataFile("history.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, &h.history)
}

func (h *HydrapotionApp) saveHistory() {
	data, _ := json.MarshalIndent(h.history, "", " ")
	os.WriteFile(h.dataFile("history.json"), data, 0644)
}

func (h *HydrapotionApp) todayStr() string {
	return time.Now().Format("2006-01-02")
}

func (h *HydrapotionApp) resetIfNewDay() {
	today := h.todayStr()
	if h.settings.LastResetDate != today {
		h.settings.TodayConsumed = 0
		h.settings.LastResetDate = today
		h.settings.CurrentMood = MoodNeutral
		h.saveSettings()
	}
}

func (h *HydrapotionApp) CalculateGoal() int {
	baseGoal := h.settings.Weight * 35

	heightBonus := 0
	if h.settings.Height > 150 {
		heightBonus = (h.settings.Height - 150) * 5
	}

	moodBonus := 0
	switch h.settings.CurrentMood {
	case MoodTense:
		moodBonus = 400
	case MoodLow:
		moodBonus = 200
	}

	weatherBonus := 0
	if h.weather != nil {
		switch {
		case h.weather.Temperature >= 35:
			weatherBonus = 600
		case h.weather.Temperature >= 30:
			weatherBonus = 400
		case h.weather.Temperature >= 25:
			weatherBonus = 200
		}
		if h.weather.Humidity > 60 {
			weatherBonus += 100
		}
	}

	return baseGoal + heightBonus + moodBonus + weatherBonus
}

func (h *HydrapotionApp) AddWater(ml int) {
	h.mu.Lock()
	h.resetIfNewDay()
	h.settings.TodayConsumed += ml

	today := h.todayStr()
	found := false
	for i, entry := range h.history {
		if entry.Date == today {
			h.history[i].Ml += ml
			found = true
			break
		}
	}
	if !found {
		h.history = append(h.history, HistoryEntry{Date: today, Ml: ml})
	}

	h.saveSettings()
	h.saveHistory()
	h.mu.Unlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}

	h.startReminderTimer()
}

func (h *HydrapotionApp) SetMood(mood Mood) {
	h.mu.Lock()
	h.settings.CurrentMood = mood
	h.saveSettings()

	today := h.todayStr()
	for i := range h.history {
		if h.history[i].Date == today {
			h.history[i].Mood = mood.String()
			break
		}
	}
	h.saveHistory()
	h.mu.Unlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}
}

func (h *HydrapotionApp) SetWeight(weight int) {
	h.mu.Lock()
	h.settings.Weight = weight
	h.saveSettings()
	h.mu.Unlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}
}

func (h *HydrapotionApp) SetHeight(height int) {
	h.mu.Lock()
	h.settings.Height = height
	h.saveSettings()
	h.mu.Unlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}
}

func (h *HydrapotionApp) SetReminderInterval(seconds int) {
	h.mu.Lock()
	h.settings.ReminderInterval = seconds
	h.saveSettings()
	h.mu.Unlock()
}

func (h *HydrapotionApp) SetLocation(name string, lat, lon float64) {
	h.mu.Lock()
	h.settings.LocationName = name
	h.settings.LocationLat = lat
	h.settings.LocationLon = lon
	h.saveSettings()
	h.mu.Unlock()

	go h.fetchWeather()

	if h.onUpdate != nil {
		fyne.Do(func() { h.onUpdate() })
	}
}

func (h *HydrapotionApp) GetProgress() (consumed int, goal int, percent float64, weather *WeatherData) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.resetIfNewDay()

	consumed = h.settings.TodayConsumed
	goal = h.CalculateGoal()
	if goal > 0 {
		percent = float64(consumed) / float64(goal) * 100
		if percent > 100 {
			percent = 100
		}
	}
	weather = h.weather
	return
}

func (h *HydrapotionApp) GetWeeklyData() []HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := []HistoryEntry{}
	now := time.Now()
	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		found := false
		for _, entry := range h.history {
			if entry.Date == date {
				result = append(result, entry)
				found = true
				break
			}
		}
		if !found {
			result = append(result, HistoryEntry{Date: date, Ml: 0})
		}
	}
	return result
}

func searchLocation(query string) ([]GeoResult, error) {
	encoded := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=5&language=es&format=json", encoded)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var geo GeoResponse
	if err := json.Unmarshal(body, &geo); err != nil {
		return nil, err
	}

	return geo.Results, nil
}

func (h *HydrapotionApp) fetchWeather() {
	if h.settings.LocationLat == 0 && h.settings.LocationLon == 0 {
		return
	}

	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m",
		h.settings.LocationLat, h.settings.LocationLon,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var weatherResp WeatherResponse
	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return
	}

	h.mu.Lock()
	h.weather = &WeatherData{
		Temperature: weatherResp.Current.Temperature,
		Humidity:    weatherResp.Current.Humidity,
	}
	h.mu.Unlock()

	if h.onUpdate != nil {
		fyne.Do(func() { h.onUpdate() })
	}
}

func (h *HydrapotionApp) startReminderTimer() {
	if h.reminder != nil {
		h.reminder.Stop()
	}

	duration := time.Duration(h.settings.ReminderInterval) * time.Second
	h.reminder = time.AfterFunc(duration, func() {
		fyne.Do(func() { h.showReminderPopup() })
	})
}

func (h *HydrapotionApp) showReminderPopup() {
	notifySend("Bebe Agua!", "Es hora de hidratarte!")

	if h.reminderWin != nil {
		ui := h.CreateReminderUI()
		h.reminderWin.SetContent(ui)
		h.reminderWin.Show()
		h.reminderWin.RequestFocus()
		return
	}

	h.reminderWin = h.app.NewWindow("Recordatorio")
	h.reminderWin.Resize(fyne.NewSize(350, 400))

	ui := h.CreateReminderUI()
	h.reminderWin.SetContent(ui)
	h.reminderWin.Show()
	playSound("popup_appear.wav")
}

func (h *HydrapotionApp) CreateReminderUI() fyne.CanvasObject {
	consumed, goal, percent, _ := h.GetProgress()

	title := widget.NewLabelWithStyle("Bebe Agua!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	progressText := widget.NewLabel(fmt.Sprintf("Hoy: %d / %d ml (%.0f%%)", consumed, goal, percent))
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.SetValue(percent)

	addWaterContainer := container.NewGridWithColumns(3,
		widget.NewButton("150ml", func() {
			h.AddWater(150)
			playSound("add_water.wav")
			h.closeReminderWindow()
		}),
		widget.NewButton("250ml", func() {
			h.AddWater(250)
			playSound("add_water.wav")
			h.closeReminderWindow()
		}),
		widget.NewButton("500ml", func() {
			h.AddWater(500)
			playSound("add_water.wav")
			h.closeReminderWindow()
		}),
	)

	snoozeContainer := container.NewGridWithColumns(3,
		widget.NewButton("5 min", func() {
			h.snoozeReminder(5)
		}),
		widget.NewButton("15 min", func() {
			h.snoozeReminder(15)
		}),
		widget.NewButton("30 min", func() {
			h.snoozeReminder(30)
		}),
	)

	closeBtn := widget.NewButton("Cerrar", func() {
		h.closeReminderWindow()
	})

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		progressText,
		progressBar,
		widget.NewSeparator(),
		widget.NewLabel("Agregar agua:"),
		addWaterContainer,
		widget.NewSeparator(),
		widget.NewLabel("Posponer:"),
		snoozeContainer,
		layout.NewSpacer(),
		closeBtn,
	)
}

func (h *HydrapotionApp) snoozeReminder(minutes int) {
	h.closeReminderWindow()
	if h.reminder != nil {
		h.reminder.Stop()
	}
	h.reminder = time.AfterFunc(time.Duration(minutes)*time.Minute, func() {
		fyne.Do(func() { h.showReminderPopup() })
	})
}

func (h *HydrapotionApp) closeReminderWindow() {
	if h.reminderWin != nil {
		h.reminderWin.Hide()
	}
}

func notifySend(title, message string) {
	cmd := exec.Command("notify-send", "-i", "water", title, message)
	cmd.Run()
}

var appSoundDir string

func initSounds() {
	home, _ := os.UserHomeDir()
	appSoundDir = filepath.Join(home, ".hydrapotion", "sounds")
	os.MkdirAll(appSoundDir, 0755)

	for _, name := range []string{"popup_appear.wav", "add_water.wav"} {
		src := filepath.Join(home, "Descargas", "Hydrapotion", name)
		dst := filepath.Join(appSoundDir, name)
		if _, err := os.Stat(src); err == nil {
			copyFile(src, dst)
		}
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func playSound(soundPath string) {
	if appSoundDir == "" {
		initSounds()
	}
	cmd := exec.Command("paplay", filepath.Join(appSoundDir, soundPath))
	cmd.Run()
}

// CircularProgress draws a ring chart with center text.
type CircularProgress struct {
	widget.BaseWidget
	Percent     float64
	Consumed    int
	Goal        int
	StrokeWidth float32
}

func NewCircularProgress() *CircularProgress {
	c := &CircularProgress{StrokeWidth: 14}
	c.ExtendBaseWidget(c)
	return c
}

func (c *CircularProgress) CreateRenderer() fyne.WidgetRenderer {
	rast := canvas.NewRaster(c.draw)
	pct := canvas.NewText("0%", theme.ForegroundColor())
	pct.Alignment = fyne.TextAlignCenter
	pct.TextSize = 28
	pct.TextStyle = fyne.TextStyle{Bold: true}

	det := canvas.NewText("0 / 0 ml", theme.ForegroundColor())
	det.Alignment = fyne.TextAlignCenter
	det.TextSize = 13

	return &circleRender{prog: c, rast: rast, pct: pct, det: det}
}

type circleRender struct {
	prog *CircularProgress
	rast *canvas.Raster
	pct  *canvas.Text
	det  *canvas.Text
}

func (r *circleRender) Layout(s fyne.Size) {
	d := s.Width
	if s.Height < d {
		d = s.Height
	}
	r.rast.Resize(fyne.NewSize(d, d))
	r.rast.Move(fyne.NewPos((s.Width-d)/2, 0))

	r.pct.Resize(fyne.NewSize(d, 36))
	r.pct.Move(fyne.NewPos((s.Width-d)/2, (d-50)/2))

	r.det.Resize(fyne.NewSize(d, 20))
	r.det.Move(fyne.NewPos((s.Width-d)/2, (d-50)/2+40))
}

func (r *circleRender) MinSize() fyne.Size { return fyne.NewSize(200, 200) }

func (r *circleRender) Refresh() {
	canvas.Refresh(r.rast)
	r.pct.Text = fmt.Sprintf("%.0f%%", r.prog.Percent)
	r.det.Text = fmt.Sprintf("%d / %d ml", r.prog.Consumed, r.prog.Goal)
	canvas.Refresh(r.pct)
	canvas.Refresh(r.det)
}

func (r *circleRender) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.rast, r.pct, r.det} }
func (r *circleRender) Destroy()                      {}

func (c *CircularProgress) draw(w, h int) image.Image {
	size := w
	if h < size {
		size = h
	}
	if size < 10 {
		size = 10
	}
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx := float64(size) / 2
	sw := float64(c.StrokeWidth)
	rad := cx - sw/2
	fgC := color.RGBA{64, 169, 255, 255}
	doneC := color.RGBA{46, 204, 113, 255}
	ringC := color.RGBA{35, 35, 45, 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cx
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < rad-sw/2 || dist > rad+sw/2 {
				continue
			}
			a := math.Atan2(dy, dx)
			if a < 0 {
				a += 2 * math.Pi
			}
			a += math.Pi / 2
			if a >= 2*math.Pi {
				a -= 2 * math.Pi
			}
			pa := c.Percent / 100.0 * 2 * math.Pi
			if a <= pa {
				if c.Percent >= 100 {
					img.Set(x, y, doneC)
				} else {
					img.Set(x, y, fgC)
				}
			} else {
				img.Set(x, y, ringC)
			}
		}
	}
	return img
}

// WaterCard is a tappable card with pixel-drawn icon and ml label.
type WaterCard struct {
	widget.BaseWidget
	ml      int
	onClick func()
	hovered bool
}

func NewWaterCard(ml int, onClick func()) *WaterCard {
	w := &WaterCard{ml: ml, onClick: onClick}
	w.ExtendBaseWidget(w)
	return w
}

func (c *WaterCard) Tapped(*fyne.PointEvent) { c.onClick() }

func (c *WaterCard) MouseIn(*desktop.MouseEvent) {
	c.hovered = true
	canvas.Refresh(c)
}

func (c *WaterCard) MouseOut() {
	c.hovered = false
	canvas.Refresh(c)
}

func (c *WaterCard) MouseMoved(*desktop.MouseEvent) {}

func (c *WaterCard) CreateRenderer() fyne.WidgetRenderer {
	icon := canvas.NewImageFromImage(glassIcon(c.ml, 48, 64, theme.PrimaryColor()))
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(36, 48))

	ml := canvas.NewText(fmt.Sprintf("%dml", c.ml), theme.ForegroundColor())
	ml.Alignment = fyne.TextAlignCenter
	ml.TextSize = c.Theme().Size(theme.SizeNameText)
	ml.TextStyle = fyne.TextStyle{Bold: true}

	bg := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	bg.StrokeColor = theme.Color(theme.ColorNameInputBorder)
	bg.StrokeWidth = 1
	bg.CornerRadius = c.Theme().Size(theme.SizeNameButtonRadius)

	wrapped := container.NewMax(
		bg,
		container.NewCenter(
			container.NewVBox(
				layout.NewSpacer(),
				icon,
				ml,
				layout.NewSpacer(),
			),
		),
	)

	return &cardRender{
		card: c, bg: bg,
		icon: icon, ml: ml,
		wrapped: wrapped,
	}
}

type cardRender struct {
	card    *WaterCard
	bg      *canvas.Rectangle
	icon    *canvas.Image
	ml      *canvas.Text
	wrapped *fyne.Container
}

func (r *cardRender) Layout(s fyne.Size) {
	r.wrapped.Resize(s)
	r.wrapped.Move(fyne.NewPos(0, 0))
}

func (r *cardRender) MinSize() fyne.Size { return fyne.NewSize(100, 100) }

func (r *cardRender) Refresh() {
	th := r.card.Theme()
	if r.card.hovered {
		r.bg.FillColor = theme.PrimaryColor()
		r.bg.CornerRadius = th.Size(theme.SizeNameButtonRadius)
		r.icon.Image = glassIcon(r.card.ml, 48, 64, theme.BackgroundColor())
		r.ml.Color = theme.BackgroundColor()
	} else {
		r.bg.FillColor = theme.Color(theme.ColorNameButton)
		r.bg.CornerRadius = th.Size(theme.SizeNameButtonRadius)
		r.icon.Image = glassIcon(r.card.ml, 48, 64, theme.PrimaryColor())
		r.ml.Color = theme.ForegroundColor()
	}
	canvas.Refresh(r.bg)
	canvas.Refresh(r.icon)
	canvas.Refresh(r.ml)
}

func (r *cardRender) Objects() []fyne.CanvasObject { return []fyne.CanvasObject{r.wrapped} }
func (r *cardRender) Destroy()                      {}

// glassIcon renders a pixel icon of a glass/bottle for the given ml.
func glassIcon(ml int, w, h int, clr color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	switch ml {
	case 150:
		drawShot(img, clr)
	case 250:
		drawPint(img, clr)
	case 500:
		drawBottle(img, clr)
	}
	return img
}

func drawShot(img *image.RGBA, clr color.Color) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	cx, cy := w/2, h/2
	topW := w * 4 / 7
	botW := w * 3 / 7
	bodyH := h * 3 / 5
	topY := cy - bodyH/2

	for y := topY; y < topY+bodyH; y++ {
		t := float64(y-topY) / float64(bodyH)
		rowW := int(float64(topW)*(1-t) + float64(botW)*t)
		sx := cx - rowW/2
		for x := sx; x < sx+rowW; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, clr)
			}
		}
	}

	baseW := w * 2 / 5
	baseH := h / 7
	for y := h - baseH; y < h; y++ {
		sx := cx - baseW/2
		for x := sx; x < sx+baseW; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, clr)
			}
		}
	}

	lig := lerpColor(clr, color.RGBA{255, 255, 255, 255}, 0.15)
	for y := topY + bodyH/4; y < topY+bodyH*3/5; y++ {
		t := float64(y-(topY+bodyH/4)) / float64(bodyH*3/5-bodyH/4)
		rowW := int(float64(topW)*(1-t) + float64(botW)*t)
		sx := cx - rowW/6
		ew := rowW / 3
		for x := sx; x < sx+ew; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, lig)
			}
		}
	}
}

func drawPint(img *image.RGBA, clr color.Color) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	cx, cy := w/2, h/2
	topW := w * 4 / 6
	botW := w * 3 / 6
	bodyH := h * 4 / 5
	topY := cy - bodyH/2 + h/10

	for y := topY; y < topY+bodyH; y++ {
		t := float64(y-topY) / float64(bodyH)
		rowW := int(float64(topW)*(1-t) + float64(botW)*t)
		sx := cx - rowW/2
		for x := sx; x < sx+rowW; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, clr)
			}
		}
	}

	baseW := w * 2 / 5
	baseH := h / 8
	for y := h - baseH; y < h; y++ {
		sx := cx - baseW/2
		for x := sx; x < sx+baseW; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, clr)
			}
		}
	}

	lig := lerpColor(clr, color.RGBA{255, 255, 255, 255}, 0.15)
	for y := topY + bodyH/4; y < topY+bodyH*3/5; y++ {
		t := float64(y-(topY+bodyH/4)) / float64(bodyH*3/5-bodyH/4)
		rowW := int(float64(topW)*(1-t) + float64(botW)*t)
		sx := cx - rowW/5
		ew := rowW / 3
		for x := sx; x < sx+ew; x++ {
			if x >= 0 && x < w {
				img.Set(x, y, lig)
			}
		}
	}
}

func drawBottle(img *image.RGBA, clr color.Color) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	cx := w / 2

	neckW := w * 2 / 7
	neckH := h * 2 / 7
	bodyW := w * 4 / 7
	bodyH := h * 5 / 7
	capH := h / 14

	yOff := h / 14

	for y := yOff; y < yOff+capH; y++ {
		sx := cx - neckW/2 - 1
		ex := cx + neckW/2 + 1
		for x := sx; x <= ex && x < w; x++ {
			if x >= 0 {
				img.Set(x, y, clr)
			}
		}
	}

	for y := yOff + capH; y < yOff+capH+neckH; y++ {
		sx := cx - neckW/2
		for x := sx; x < sx+neckW && x < w; x++ {
			if x >= 0 {
				img.Set(x, y, clr)
			}
		}
	}

	for y := yOff + capH + neckH; y < h; y++ {
		t := float64(y-(yOff+capH+neckH)) / float64(h-(yOff+capH+neckH))
		rowW := int(float64(neckW+2)*(1-t) + float64(bodyW)*t)
		if rowW > bodyW {
			rowW = bodyW
		}
		sx := cx - rowW/2
		for x := sx; x < sx+rowW && x < w; x++ {
			if x >= 0 {
				img.Set(x, y, clr)
			}
		}
	}

	lig := lerpColor(clr, color.RGBA{255, 255, 255, 255}, 0.15)
	for y := yOff + capH + neckH + bodyH/5; y < yOff+capH+neckH+bodyH*3/5; y++ {
		t := float64(y-(yOff+capH+neckH+bodyH/5)) / float64(bodyH*3/5-bodyH/5)
		rowW := int(float64(bodyW/3)*(1-t) + float64(bodyW/3)*t)
		sx := cx - bodyW/4
		for x := sx; x < sx+rowW && x < w && x >= 0; x++ {
			img.Set(x, y, lig)
		}
	}
}

func lerpColor(a, b color.Color, t float64) color.Color {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return color.RGBA{
		R: uint8(float64(ar>>8)*(1-t) + float64(br>>8)*t),
		G: uint8(float64(ag>>8)*(1-t) + float64(bg>>8)*t),
		B: uint8(float64(ab>>8)*(1-t) + float64(bb>>8)*t),
		A: uint8(float64(aa>>8)*(1-t) + float64(ba>>8)*t),
	}
}

// ----------------------------------------------------------------
// WaterChart — daily bar chart, medical-grade
// ----------------------------------------------------------------
type WaterChart struct {
	widget.BaseWidget
	data []HistoryEntry
	goal int
	avg  int
}

func NewWaterChart() *WaterChart {
	w := &WaterChart{}
	w.ExtendBaseWidget(w)
	return w
}

func (ch *WaterChart) SetData(data []HistoryEntry, goal int) {
	ch.data = data
	ch.goal = goal
	sum := 0
	for i := range ch.data {
		if ch.data[i].Ml < 0 {
			ch.data[i].Ml = 0
		}
		sum += ch.data[i].Ml
	}
	if len(ch.data) > 0 {
		ch.avg = sum / len(ch.data)
	}
	canvas.Refresh(ch)
}

func (ch *WaterChart) CreateRenderer() fyne.WidgetRenderer {
	rast := canvas.NewRaster(ch.draw)

	dayLbls := make([]*canvas.Text, 7)
	for i := range dayLbls {
		t := canvas.NewText("", color.RGBA{180, 180, 195, 255})
		t.TextSize = 11
		t.Alignment = fyne.TextAlignCenter
		dayLbls[i] = t
	}

	y0 := canvas.NewText("0", color.RGBA{180, 180, 195, 255})
	y0.TextSize = 11

	yGoal := canvas.NewText("Meta", color.RGBA{220, 80, 80, 255})
	yGoal.TextSize = 11

	avgLbl := canvas.NewText("Prom", color.RGBA{240, 200, 60, 255})
	avgLbl.TextSize = 11

	objs := []fyne.CanvasObject{rast, y0, yGoal, avgLbl}
	for _, d := range dayLbls {
		objs = append(objs, d)
	}

	return &chartRender{
		chart: ch, rast: rast,
		y0: y0, yGoal: yGoal, avgLbl: avgLbl,
		days: dayLbls, objs: objs,
	}
}

type chartRender struct {
	chart  *WaterChart
	rast   *canvas.Raster
	y0     *canvas.Text
	yGoal  *canvas.Text
	avgLbl *canvas.Text
	days   []*canvas.Text
	objs   []fyne.CanvasObject
}

func (r *chartRender) Layout(s fyne.Size) {
	const left, right, top, bottom = 50, 15, 20, 35
	chartW := s.Width - left - right
	chartH := s.Height - top - bottom
	if chartW < 10 || chartH < 10 {
		return
	}

	r.rast.Resize(s)

	n := len(r.chart.data)
	if n == 0 {
		return
	}

	maxVal := r.chart.goal
	for _, e := range r.chart.data {
		if e.Ml > maxVal {
			maxVal = e.Ml
		}
	}
	if maxVal <= 0 {
		maxVal = 1000
	}

	barW := float32(chartW) / float32(n)
	const maxBarW float32 = 36
	if barW > maxBarW {
		barW = maxBarW
	}
	gap := (float32(chartW) - barW*float32(n)) / (float32(n) + 1)
	if gap < 3 {
		gap = 3
		barW = (float32(chartW) - gap*(float32(n)+1)) / float32(n)
	}

		// Y=0 label
	r.y0.Move(fyne.NewPos(4, s.Height-bottom-9))
	r.y0.Resize(fyne.NewSize(left-8, 14))

	// Goal label + position
	if r.chart.goal > 0 && maxVal > 0 {
		r.yGoal.Text = fmt.Sprintf("%dml", r.chart.goal)
		goalY := float32(float64(top) + float64(chartH)*(1-float64(r.chart.goal)/float64(maxVal)))
		r.yGoal.Move(fyne.NewPos(4, goalY-7))
		r.yGoal.Resize(fyne.NewSize(left-8, 14))
		canvas.Refresh(r.yGoal)
	}

	// Avg label
	if r.chart.avg > 0 && maxVal > 0 {
		r.avgLbl.Text = fmt.Sprintf("%dml", r.chart.avg)
		avgY := float32(float64(top) + float64(chartH)*(1-float64(r.chart.avg)/float64(maxVal)))
		r.avgLbl.Move(fyne.NewPos(4, avgY-7))
		r.avgLbl.Resize(fyne.NewSize(left-8, 14))
		canvas.Refresh(r.avgLbl)
	}

	dayNames := []string{"Dom", "Lun", "Mar", "Mie", "Jue", "Vie", "Sab"}
	for i, e := range r.chart.data {
		if i >= len(r.days) {
			break
		}
		x := float32(left) + gap + float32(i)*(barW+gap)
		t, err := time.Parse("2006-01-02", e.Date)
		name := dayNames[t.Weekday()]
		if err != nil {
			name = "?"
		}
		r.days[i].Text = name
		r.days[i].Move(fyne.NewPos(x+barW/2-14, s.Height-bottom+4))
		r.days[i].Resize(fyne.NewSize(28, 14))
		canvas.Refresh(r.days[i])
	}
}

func (r *chartRender) MinSize() fyne.Size { return fyne.NewSize(320, 160) }

func (r *chartRender) Refresh() {
	canvas.Refresh(r.rast)
}

func (r *chartRender) Objects() []fyne.CanvasObject { return r.objs }

func (r *chartRender) Destroy() {}

func (ch *WaterChart) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{25, 25, 38, 255}
	gridC := color.RGBA{50, 50, 70, 255}
	barC := color.RGBA{64, 169, 255, 200}
	doneC := color.RGBA{46, 204, 113, 200}
	goalLineC := color.RGBA{220, 80, 80, 200}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bg)
		}
	}

	const left, right, top, bottom = 50, 15, 20, 35
	chartW := w - left - right
	chartH := h - top - bottom
	if chartW < 10 || chartH < 10 {
		return img
	}

	n := len(ch.data)
	if n == 0 {
		return img
	}

	maxVal := ch.goal
	for _, e := range ch.data {
		if e.Ml > maxVal {
			maxVal = e.Ml
		}
	}
	if maxVal <= 0 {
		maxVal = 1000
	}

	// Baseline
	for x := left; x < left+chartW; x++ {
		img.Set(x, top+chartH, gridC)
	}

	// Goal line (dashed)
	if ch.goal > 0 && maxVal > 0 {
		goalY := top + chartH - int(float64(ch.goal)/float64(maxVal)*float64(chartH))
		if goalY >= top && goalY <= top+chartH {
			for x := left; x < left+chartW; x += 8 {
				end := x + 4
				if end > left+chartW {
					end = left + chartW
				}
				for dx := x; dx < end; dx++ {
					img.Set(dx, goalY, goalLineC)
				}
			}
		}
	}

	// Avg line (dotted amber)
	if ch.avg > 0 && maxVal > 0 {
		avgY := top + chartH - int(float64(ch.avg)/float64(maxVal)*float64(chartH))
		if avgY >= top && avgY <= top+chartH {
			for x := left; x < left+chartW; x += 6 {
				img.Set(x, avgY, color.RGBA{240, 200, 60, 180})
			}
		}
	}

	// Bar dimensions
	barW := chartW / n
	const maxBarW = 36
	if barW > maxBarW {
		barW = maxBarW
	}
	gap := (chartW - barW*n) / (n + 1)
	if gap < 3 {
		gap = 3
		barW = (chartW - gap*(n+1)) / n
	}

	for i, e := range ch.data {
		x := left + gap + i*(barW+gap)
		barH := int(float64(e.Ml) / float64(maxVal) * float64(chartH))
		if barH < 0 {
			barH = 0
		}
		barY := top + chartH - barH

		clr := barC
		if e.Ml >= ch.goal && ch.goal > 0 {
			clr = doneC
		}

		for by := barY; by < top+chartH; by++ {
			for bx := x; bx < x+barW; bx++ {
				if bx >= left && bx < left+chartW {
					img.Set(bx, by, clr)
				}
			}
		}
	}

	return img
}

// ----------------------------------------------------------------
// Main UI
// ----------------------------------------------------------------
func (h *HydrapotionApp) CreateMainUI() fyne.CanvasObject {
	circ := NewCircularProgress()

	c1 := NewWaterCard(150, func() {
		h.AddWater(150)
		playSound("add_water.wav")
	})
	c2 := NewWaterCard(250, func() {
		h.AddWater(250)
		playSound("add_water.wav")
	})
	c3 := NewWaterCard(500, func() {
		h.AddWater(500)
		playSound("add_water.wav")
	})
	grid := container.NewGridWrap(fyne.NewSize(100, 100), c1, c2, c3)

	moodOpts := []string{"😊 Bien", "😐 Neutral", "😞 Bajo", "😰 Tenso"}
	moodSel := widget.NewSelect(moodOpts, func(s string) {
		var m Mood
		// Strip emoji prefix to get mood name
		parts := strings.SplitN(s, " ", 2)
		name := parts[1]
		switch name {
		case "Bien":
			m = MoodWell
		case "Neutral":
			m = MoodNeutral
		case "Bajo":
			m = MoodLow
		case "Tenso":
			m = MoodTense
		}
		h.SetMood(m)
	})
	for _, opt := range moodOpts {
		if strings.Contains(opt, h.settings.CurrentMood.String()) {
			moodSel.SetSelected(opt)
			break
		}
	}

	moodRow := container.NewHBox(
		widget.NewLabel("Estado mental:"),
		layout.NewSpacer(),
		moodSel,
	)

	gear := widget.NewButtonWithIcon("", theme.MenuIcon(), func() { h.showSettings() })

	titleRow := container.NewHBox(
		widget.NewLabelWithStyle("Hydrapotion", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		gear,
	)

	secTitle := func(t string) fyne.CanvasObject {
		l := widget.NewLabelWithStyle(t, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
		l.Importance = widget.MediumImportance
		return l
	}

	chart := NewWaterChart()

	h.onUpdate = func() {
		c, g, p, _ := h.GetProgress()
		circ.Percent = p
		circ.Consumed = c
		circ.Goal = g
		circ.Refresh()
		chart.SetData(h.GetWeeklyData(), h.CalculateGoal())
	}
	h.onUpdate()

	body := container.NewVBox(
		circ,
		container.NewCenter(secTitle("¿Cuánto vas a tomar?")),
		container.NewCenter(container.NewPadded(grid)),
		container.NewPadded(chart),
	)

	return container.NewBorder(
		container.NewPadded(titleRow),
		container.NewPadded(moodRow),
		nil, nil,
		body,
	)
}

// ----------------------------------------------------------------
// Settings dialog
// ----------------------------------------------------------------
func (h *HydrapotionApp) showSettings() {
	if h.settingsWin != nil {
		h.settingsWin.Show()
		h.settingsWin.RequestFocus()
		return
	}

	d := h.app.NewWindow("Configuración")
	d.Resize(fyne.NewSize(360, 440))
	h.settingsWin = d

	d.SetCloseIntercept(func() {
		d.Hide()
	})

	wEntry := widget.NewEntry()
	wEntry.SetPlaceHolder("Peso en kg")
	wEntry.SetText(fmt.Sprintf("%d", h.settings.Weight))
	wBtn := widget.NewButton("Guardar", func() {
		var v int
		fmt.Sscanf(wEntry.Text, "%d", &v)
		if v > 0 {
			h.SetWeight(v)
		}
	})

	hEntry := widget.NewEntry()
	hEntry.SetPlaceHolder("Altura en cm")
	hEntry.SetText(fmt.Sprintf("%d", h.settings.Height))
	hBtn := widget.NewButton("Guardar", func() {
		var v int
		fmt.Sscanf(hEntry.Text, "%d", &v)
		if v > 0 {
			h.SetHeight(v)
		}
	})

	curMin := h.settings.ReminderInterval / 60
	ivSel := widget.NewSelect(
		[]string{"5 min", "10 min", "15 min", "30 min", "45 min", "60 min"},
		func(s string) {
			var m int
			fmt.Sscanf(s, "%d", &m)
			h.SetReminderInterval(m * 60)
		},
	)
	ivSel.SetSelected(fmt.Sprintf("%d min", curMin))

	locLabel := widget.NewLabel("No configurada")
	locLabel.Wrapping = fyne.TextWrapWord
	if h.settings.LocationName != "" {
		t := h.settings.LocationName
		if h.weather != nil {
			t += fmt.Sprintf(" — %.1f°C %d%% HR", h.weather.Temperature, int(h.weather.Humidity))
		}
		locLabel.SetText(t)
	}
	locBtn := widget.NewButton("Buscar ubicación", func() {
		h.showLocationDialog()
		d.Hide()
	})

	closeBtn := widget.NewButton("Cerrar", func() { d.Hide() })

	sec := func(t string) fyne.CanvasObject {
		l := widget.NewLabelWithStyle(t, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		l.Importance = widget.MediumImportance
		return l
	}

	c := container.NewVBox(
		sec("Peso"),
		container.NewHBox(wEntry, wBtn),
		widget.NewSeparator(),
		sec("Altura"),
		container.NewHBox(hEntry, hBtn),
		widget.NewSeparator(),
		sec("Recordatorio"),
		ivSel,
		widget.NewSeparator(),
		sec("Ubicación"),
		locLabel,
		locBtn,
		widget.NewSeparator(),
		sec("Exportar datos"),
		widget.NewButton("Exportar JSON", func() { h.exportJSON(d) }),
		widget.NewButton("Exportar CSV", func() { h.exportCSV(d) }),
		layout.NewSpacer(),
		widget.NewSeparator(),
		closeBtn,
	)

	d.SetContent(container.NewPadded(c))
	d.Show()
}

// ----------------------------------------------------------------
// Location search dialog
// ----------------------------------------------------------------
func (h *HydrapotionApp) showLocationDialog() {
	d := h.app.NewWindow("Buscar ubicación")
	d.Resize(fyne.NewSize(380, 320))

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Ciudad, país...")
	status := widget.NewLabel("")
	resBox := container.NewVBox()

	search := widget.NewButton("Buscar", func() {
		if entry.Text == "" {
			return
		}
		status.SetText("Buscando...")
		resBox.RemoveAll()
		go func() {
			results, err := searchLocation(entry.Text)
			fyne.Do(func() {
				if err != nil || len(results) == 0 {
					status.SetText("Sin resultados")
					return
				}
				status.SetText(fmt.Sprintf("%d resultado(s):", len(results)))
				for _, r := range results {
					rr := r
					btn := widget.NewButton(
						fmt.Sprintf("%s, %s", rr.Name, rr.Country),
						func() {
							h.SetLocation(
								fmt.Sprintf("%s, %s", rr.Name, rr.Country),
								rr.Latitude, rr.Longitude,
							)
							d.Close()
						},
					)
					resBox.Add(btn)
				}
			})
		}()
	})

	d.SetContent(container.NewPadded(
		container.NewVBox(
			entry, search, widget.NewSeparator(),
			status, resBox,
			layout.NewSpacer(),
			widget.NewButton("Cancelar", func() { d.Close() }),
		),
	))
	d.Show()
}

// ----------------------------------------------------------------
// Export
// ----------------------------------------------------------------
func (h *HydrapotionApp) exportJSON(w fyne.Window) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.MarshalIndent(h.history, "", "  ")
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		defer wc.Close()
		wc.Write(data)
		dialog.ShowInformation("Exportar", "JSON guardado", w)
	}, w)
}

func (h *HydrapotionApp) exportCSV(w fyne.Window) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString("Fecha,ml,Estado de animo\n")
	for _, e := range h.history {
		mood := e.Mood
		if mood == "" {
			mood = "-"
		}
		buf.WriteString(fmt.Sprintf("%s,%d,%s\n", e.Date, e.Ml, mood))
	}

	dialog.ShowFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil || wc == nil {
			return
		}
		defer wc.Close()
		wc.Write(buf.Bytes())
		dialog.ShowInformation("Exportar", "CSV guardado", w)
	}, w)
}

// ----------------------------------------------------------------
// System tray
// ----------------------------------------------------------------
func (h *HydrapotionApp) setupSystemTray() {
	if desk, ok := h.app.(desktop.App); ok {
		m := fyne.NewMenu("Hydrapotion",
			fyne.NewMenuItem("150ml", func() { h.AddWater(150) }),
			fyne.NewMenuItem("250ml", func() { h.AddWater(250) }),
			fyne.NewMenuItem("500ml", func() { h.AddWater(500) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Configuración", func() { h.showSettings() }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Mostrar ventana", func() {
				if h.mainWindow != nil {
					h.mainWindow.Show()
					h.mainWindow.RequestFocus()
				}
			}),
			fyne.NewMenuItem("Salir", func() { h.app.Quit() }),
		)
		desk.SetSystemTrayMenu(m)
	}
}

// ----------------------------------------------------------------
// Entry point
// ----------------------------------------------------------------
func main() {
	a := app.NewWithID("hydrapotion")
	w := a.NewWindow("Hydrapotion")
	a.Settings().SetTheme(theme.DarkTheme())

	h := NewHydrapotionApp(a)
	h.mainWindow = w
	h.setupSystemTray()

	w.SetContent(container.NewPadded(h.CreateMainUI()))
	w.Resize(fyne.NewSize(360, 540))
	w.SetFixedSize(true)
	w.SetCloseIntercept(func() { w.Hide() })

	h.startReminderTimer()
	w.ShowAndRun()
}
