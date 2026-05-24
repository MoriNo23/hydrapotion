package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
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

// ----------------------------------------------------------------
// i18n
// ----------------------------------------------------------------
var currentLang string = "es"

var translations = map[string]map[string]string{
	"how_much":         {"es": "¿Cuánto vas a tomar?", "en": "How much will you drink?", "zh": "你要喝多少？"},
	"mood":             {"es": "Estado mental:", "en": "Mood:", "zh": "情绪状态："},
	"mood_well":        {"es": "Bien", "en": "Good", "zh": "好"},
	"mood_neutral":     {"es": "Neutral", "en": "Neutral", "zh": "一般"},
	"mood_low":         {"es": "Bajo", "en": "Low", "zh": "低落"},
	"mood_tense":       {"es": "Tenso", "en": "Tense", "zh": "紧张"},
	"goal":             {"es": "Meta", "en": "Goal", "zh": "目标"},
	"avg":              {"es": "Prom", "en": "Avg", "zh": "平均"},
	"sun":              {"es": "Dom", "en": "Sun", "zh": "日"},
	"mon":              {"es": "Lun", "en": "Mon", "zh": "一"},
	"tue":              {"es": "Mar", "en": "Tue", "zh": "二"},
	"wed":              {"es": "Mie", "en": "Wed", "zh": "三"},
	"thu":              {"es": "Jue", "en": "Thu", "zh": "四"},
	"fri":              {"es": "Vie", "en": "Fri", "zh": "五"},
	"sat":              {"es": "Sab", "en": "Sat", "zh": "六"},
	"settings_title":   {"es": "Configuración", "en": "Settings", "zh": "设置"},
	"weight":           {"es": "Peso", "en": "Weight", "zh": "体重"},
	"weight_ph":        {"es": "Peso en kg", "en": "Weight in kg", "zh": "体重（公斤）"},
	"height":           {"es": "Altura", "en": "Height", "zh": "身高"},
	"height_ph":        {"es": "Altura en cm", "en": "Height in cm", "zh": "身高（厘米）"},
	"save":             {"es": "Guardar", "en": "Save", "zh": "保存"},
	"reminder":         {"es": "Recordatorio", "en": "Reminder", "zh": "提醒"},
	"min":              {"es": "min", "en": "min", "zh": "分钟"},
	"hr":               {"es": "h", "en": "h", "zh": "小时"},
	"location":         {"es": "Ubicación", "en": "Location", "zh": "位置"},
	"not_set":          {"es": "No configurada", "en": "Not set", "zh": "未设置"},
	"search_location":  {"es": "Buscar ubicación", "en": "Search location", "zh": "搜索位置"},
	"export_data":      {"es": "Exportar datos", "en": "Export data", "zh": "导出数据"},
	"export_json":      {"es": "Exportar JSON", "en": "Export JSON", "zh": "导出 JSON"},
	"export_csv":       {"es": "Exportar CSV", "en": "Export CSV", "zh": "导出 CSV"},
	"close":            {"es": "Cerrar", "en": "Close", "zh": "关闭"},
	"export":           {"es": "Exportar", "en": "Export", "zh": "导出"},
	"json_saved":       {"es": "JSON guardado", "en": "JSON saved", "zh": "JSON 已保存"},
	"csv_saved":        {"es": "CSV guardado", "en": "CSV saved", "zh": "CSV 已保存"},
	"search_title":     {"es": "Buscar ubicación", "en": "Search location", "zh": "搜索位置"},
	"city_ph":          {"es": "Ciudad, país...", "en": "City, country...", "zh": "城市，国家..."},
	"search_btn":       {"es": "Buscar", "en": "Search", "zh": "搜索"},
	"searching":        {"es": "Buscando...", "en": "Searching...", "zh": "搜索中..."},
	"no_results":       {"es": "Sin resultados", "en": "No results", "zh": "无结果"},
	"cancel":           {"es": "Cancelar", "en": "Cancel", "zh": "取消"},
	"results":          {"es": "resultado(s):", "en": "result(s):", "zh": "个结果："},
	"reminder_title":   {"es": "Recordatorio", "en": "Reminder", "zh": "提醒"},
	"drink_title":      {"es": "Bebe Agua!", "en": "Drink Water!", "zh": "喝水！"},
	"today_progress":   {"es": "Hoy: %d / %d ml (%.0f%%)", "en": "Today: %d / %d ml (%.0f%%)", "zh": "今日：%d / %d 毫升（%.0f%%）"},
	"add_water":        {"es": "Agregar agua:", "en": "Add water:", "zh": "添加水："},
	"snooze":           {"es": "Posponer:", "en": "Snooze:", "zh": "稍后提醒："},
	"close_btn":        {"es": "Cerrar", "en": "Close", "zh": "关闭"},
	"notify_title":     {"es": "Bebe Agua!", "en": "Drink Water!", "zh": "喝水！"},
	"notify_body":      {"es": "Es hora de hidratarte!", "en": "Time to hydrate!", "zh": "该喝水了！"},
	"settings_menu":    {"es": "Configuración", "en": "Settings", "zh": "设置"},
	"show_window":      {"es": "Mostrar ventana", "en": "Show window", "zh": "显示窗口"},
	"hide_tray":        {"es": "Minimizar a bandeja", "en": "Hide to tray", "zh": "最小化到托盘"},
	"quit":             {"es": "Salir", "en": "Quit", "zh": "退出"},
}

func T(key string) string {
	if m, ok := translations[key]; ok {
		if s, ok := m[currentLang]; ok {
			return s
		}
		return m["es"]
	}
	return key
}

// ----------------------------------------------------------------
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
	currentLang = h.settings.Language
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

func (h *HydrapotionApp) writeStatusFile() {
	c, g, p, _ := h.GetProgress()

	h.mu.RLock()
	mood := h.settings.CurrentMood.String()
	h.mu.RUnlock()

	data, _ := json.Marshal(map[string]interface{}{
		"consumed": c,
		"goal":     g,
		"percent":  math.Round(p*10) / 10,
		"mood":     mood,
	})
	os.WriteFile(h.dataFile("status.json"), data, 0644)
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

func (h *HydrapotionApp) dynamicInterval() time.Duration {
	base := time.Duration(h.settings.ReminderInterval) * time.Second
	if base <= 0 {
		base = 2 * time.Hour
	}

	hour := time.Now().Hour()
	var timeFactor float64
	switch {
	case hour >= 6 && hour < 12:
		timeFactor = 0.8
	case hour >= 12 && hour < 18:
		timeFactor = 0.7
	case hour >= 18 && hour < 22:
		timeFactor = 1.0
	default:
		timeFactor = 2.0
	}

	weatherFactor := 1.0
	if h.weather != nil {
		switch {
		case h.weather.Temperature >= 35:
			weatherFactor = 0.5
		case h.weather.Temperature >= 30:
			weatherFactor = 0.6
		case h.weather.Temperature >= 25:
			weatherFactor = 0.8
		}
		if h.weather.Humidity > 70 {
			weatherFactor *= 0.9
		}
	}

	factor := timeFactor * weatherFactor
	if factor < 0.5 {
		factor = 0.5
	}

	d := time.Duration(float64(base) * factor)
	if d < 30*time.Minute {
		d = 30 * time.Minute
	}
	if d > 6*time.Hour {
		d = 6 * time.Hour
	}
	return d
}

func (h *HydrapotionApp) startReminderTimer() {
	if h.reminder != nil {
		h.reminder.Stop()
	}

	duration := h.dynamicInterval()
	h.reminder = time.AfterFunc(duration, func() {
		fyne.Do(func() { h.showReminderPopup() })
	})
}

func (h *HydrapotionApp) showReminderPopup() {
	notifySend(T("notify_title"), T("notify_body"))

	if h.reminderWin != nil {
		ui := h.CreateReminderUI()
		h.reminderWin.SetContent(ui)
		h.reminderWin.Show()
		h.reminderWin.RequestFocus()
		return
	}

	h.reminderWin = h.app.NewWindow(T("reminder_title"))
	h.reminderWin.Resize(fyne.NewSize(350, 400))

	ui := h.CreateReminderUI()
	h.reminderWin.SetContent(ui)
	h.reminderWin.Show()
	playSound("popup_appear.wav")
}

func (h *HydrapotionApp) CreateReminderUI() fyne.CanvasObject {
	consumed, goal, percent, _ := h.GetProgress()

	title := widget.NewLabelWithStyle(T("drink_title"), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	progressText := widget.NewLabel(fmt.Sprintf(T("today_progress"), consumed, goal, percent))
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
		widget.NewButton("5 "+T("min"), func() {
			h.snoozeReminder(5)
		}),
		widget.NewButton("15 "+T("min"), func() {
			h.snoozeReminder(15)
		}),
		widget.NewButton("30 "+T("min"), func() {
			h.snoozeReminder(30)
		}),
	)

	closeBtn := widget.NewButton(T("close_btn"), func() {
		h.closeReminderWindow()
	})

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		progressText,
		progressBar,
		widget.NewSeparator(),
		widget.NewLabel(T("add_water")),
		addWaterContainer,
		widget.NewSeparator(),
		widget.NewLabel(T("snooze")),
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
	h.startReminderTimer()
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

	yGoal := canvas.NewText(T("goal"), color.RGBA{220, 80, 80, 255})
	yGoal.TextSize = 11

	avgLbl := canvas.NewText(T("avg"), color.RGBA{240, 200, 60, 255})
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

	dayNames := []string{T("sun"), T("mon"), T("tue"), T("wed"), T("thu"), T("fri"), T("sat")}
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

	moodOpts := []string{
		"😊 " + T("mood_well"),
		"😐 " + T("mood_neutral"),
		"😞 " + T("mood_low"),
		"😰 " + T("mood_tense"),
	}
	moodSel := widget.NewSelect(moodOpts, func(s string) {
		for i, opt := range moodOpts {
			if opt == s {
				h.SetMood(Mood(i))
				break
			}
		}
	})
	moodSel.SetSelected(moodOpts[h.settings.CurrentMood])

	moodRow := container.NewHBox(
		widget.NewLabel(T("mood")),
		layout.NewSpacer(),
		moodSel,
	)

	gear := widget.NewButtonWithIcon("", theme.MenuIcon(), func() { h.showSettings() })

	langs := []string{"es", "en", "zh"}
	langNames := []string{"🇪🇸", "🇬🇧", "🇨🇳"}
	langIdx := 0
	for i, l := range langs {
		if h.settings.Language == l {
			langIdx = i
			break
		}
	}
	langBtn := widget.NewButton(langNames[langIdx], func() {
		cur := 0
		for i, l := range langs {
			if h.settings.Language == l {
				cur = (i + 1) % len(langs)
				break
			}
		}
		h.setLang(langs[cur])
	})

	minBtn := widget.NewButton("—", func() {
		h.mainWindow.Hide()
	})

	titleRow := container.NewHBox(
		widget.NewLabelWithStyle("Hydrapotion", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		minBtn, langBtn, gear,
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
		h.writeStatusFile()
	}
	h.onUpdate()

	body := container.NewVBox(
		circ,
		container.NewCenter(secTitle(T("how_much"))),
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
// Language
// ----------------------------------------------------------------
func (h *HydrapotionApp) setLang(l string) {
	if h.settings.Language == l {
		return
	}
	h.settings.Language = l
	h.saveSettings()
	currentLang = l

	if h.settingsWin != nil {
		h.settingsWin.Hide()
	}

	h.rebuildUI()
}

func (h *HydrapotionApp) rebuildUI() {
	content := container.NewPadded(h.CreateMainUI())
	h.mainWindow.SetContent(content)
}

// ----------------------------------------------------------------
// Settings dialog
// ----------------------------------------------------------------
func (h *HydrapotionApp) buildSettingsContent(d fyne.Window) fyne.CanvasObject {
	wEntry := widget.NewEntry()
	wEntry.SetPlaceHolder(T("weight_ph"))
	wEntry.SetText(fmt.Sprintf("%d", h.settings.Weight))
	wBtn := widget.NewButton(T("save"), func() {
		var v int
		fmt.Sscanf(wEntry.Text, "%d", &v)
		if v > 0 {
			h.SetWeight(v)
		}
	})

	hEntry := widget.NewEntry()
	hEntry.SetPlaceHolder(T("height_ph"))
	hEntry.SetText(fmt.Sprintf("%d", h.settings.Height))
	hBtn := widget.NewButton(T("save"), func() {
		var v int
		fmt.Sscanf(hEntry.Text, "%d", &v)
		if v > 0 {
			h.SetHeight(v)
		}
	})

	intervalOpts := []string{"1 " + T("hr"), "1.5 " + T("hr"), "2 " + T("hr"), "2.5 " + T("hr"), "3 " + T("hr")}
	intervalVals := []int{3600, 5400, 7200, 9000, 10800}
	curSec := h.settings.ReminderInterval
	curIdx := 2 // default 2h
	for i, v := range intervalVals {
		if curSec == v {
			curIdx = i
			break
		}
	}
	ivSel := widget.NewSelect(intervalOpts, func(s string) {
		for i, opt := range intervalOpts {
			if opt == s {
				h.SetReminderInterval(intervalVals[i])
				break
			}
		}
	})
	ivSel.SetSelected(intervalOpts[curIdx])

	locLabel := widget.NewLabel(T("not_set"))
	locLabel.Wrapping = fyne.TextWrapWord
	if h.settings.LocationName != "" {
		t := h.settings.LocationName
		if h.weather != nil {
			t += fmt.Sprintf(" — %.1f°C %d%% RH", h.weather.Temperature, int(h.weather.Humidity))
		}
		locLabel.SetText(t)
	}
	locBtn := widget.NewButton(T("search_location"), func() {
		h.showLocationDialog()
	})

	closeBtn := widget.NewButton(T("close"), func() { d.Hide() })

	sec := func(t string) fyne.CanvasObject {
		l := widget.NewLabelWithStyle(t, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		l.Importance = widget.MediumImportance
		return l
	}

	return container.NewVBox(
		sec(T("weight")),
		container.NewHBox(wEntry, wBtn),
		widget.NewSeparator(),
		sec(T("height")),
		container.NewHBox(hEntry, hBtn),
		widget.NewSeparator(),
		sec(T("reminder")),
		ivSel,
		widget.NewSeparator(),
		sec(T("location")),
		locLabel,
		locBtn,
		widget.NewSeparator(),
		sec(T("export_data")),
		widget.NewButton(T("export_json"), func() { h.exportJSON(d) }),
		widget.NewButton(T("export_csv"), func() { h.exportCSV(d) }),
		layout.NewSpacer(),
		widget.NewSeparator(),
		closeBtn,
	)
}

func (h *HydrapotionApp) showSettings() {
	if h.settingsWin != nil {
		h.settingsWin.SetContent(container.NewPadded(h.buildSettingsContent(h.settingsWin)))
		h.settingsWin.Show()
		h.settingsWin.RequestFocus()
		return
	}

	d := h.app.NewWindow(T("settings_title"))
	d.Resize(fyne.NewSize(360, 440))
	h.settingsWin = d

	d.SetCloseIntercept(func() {
		d.Hide()
	})

	d.SetContent(container.NewPadded(h.buildSettingsContent(d)))
	d.Show()
}

// ----------------------------------------------------------------
// Location search dialog
// ----------------------------------------------------------------
func (h *HydrapotionApp) showLocationDialog() {
	d := h.app.NewWindow(T("search_title"))
	d.Resize(fyne.NewSize(380, 320))

	entry := widget.NewEntry()
	entry.SetPlaceHolder(T("city_ph"))
	status := widget.NewLabel("")
	resBox := container.NewVBox()

	search := widget.NewButton(T("search_btn"), func() {
		if entry.Text == "" {
			return
		}
		status.SetText(T("searching"))
		resBox.RemoveAll()
		go func() {
			results, err := searchLocation(entry.Text)
			fyne.Do(func() {
				if err != nil || len(results) == 0 {
					status.SetText(T("no_results"))
					return
				}
				status.SetText(fmt.Sprintf("%d %s", len(results), T("results")))
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
			widget.NewButton(T("cancel"), func() { d.Close() }),
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
		dialog.ShowInformation(T("export"), T("json_saved"), w)
	}, w)
}

func (h *HydrapotionApp) exportCSV(w fyne.Window) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var buf bytes.Buffer
	buf.WriteString("Date,ml,Mood\n")
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
		dialog.ShowInformation(T("export"), T("csv_saved"), w)
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
			fyne.NewMenuItem(T("settings_menu"), func() { h.showSettings() }),
			fyne.NewMenuItem(T("hide_tray"), func() { h.mainWindow.Hide() }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem(T("show_window"), func() {
				if h.mainWindow != nil {
					h.mainWindow.Show()
					h.mainWindow.RequestFocus()
				}
			}),
			fyne.NewMenuItem(T("quit"), func() { h.app.Quit() }),
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
