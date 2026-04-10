package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// --- Models ---

type Mood int

const (
	MoodWell Mood = iota
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
	Weight           int    `json:"weight"`
	TodayConsumed    int    `json:"today_consumed"`
	LastResetDate    string `json:"last_reset_date"`
	Language         string `json:"language"`
	Location         string `json:"location"`
	ReminderInterval int    `json:"reminder_interval"` // en segundos
	CurrentMood      Mood   `json:"current_mood"`
}

type HistoryEntry struct {
	Date string `json:"date"`
	Ml   int    `json:"ml"`
}

// --- App State ---

type HydrapotionApp struct {
	settings     Settings
	history      []HistoryEntry
	mu           sync.RWMutex
	dataDir      string
	reminder     *time.Timer
	onUpdate     func()
	mainWindow   fyne.Window
	reminderWin  fyne.Window
	app          fyne.App
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
			Language:         "es",
			ReminderInterval: 1800, // 30 min default
			CurrentMood:      MoodNeutral,
		},
	}

	h.loadSettings()
	h.loadHistory()
	h.resetIfNewDay()

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

	moodBonus := 0
	switch h.settings.CurrentMood {
	case MoodTense:
		moodBonus = 400
	case MoodLow:
		moodBonus = 200
	}

	return baseGoal + moodBonus
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

func (h *HydrapotionApp) GetProgress() (consumed int, goal int, percent float64) {
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

func (h *HydrapotionApp) startReminderTimer() {
	if h.reminder != nil {
		h.reminder.Stop()
	}

	// 20 segundos para pruebas
	duration := 20 * time.Second
	h.reminder = time.AfterFunc(duration, func() {
		h.showReminderPopup()
	})
}

func (h *HydrapotionApp) showReminderPopup() {
	// Notificacion del sistema
	notifySend("Bebe Agua!", "Es hora de hidratarte!")

	// Mostrar ventana popup
	if h.reminderWin != nil {
		h.reminderWin.Show()
		h.reminderWin.RequestFocus()
		return
	}

	// Crear ventana de recordatorio
	h.reminderWin = h.app.NewWindow("Recordatorio")
	h.reminderWin.Resize(fyne.NewSize(350, 400))

	ui := h.CreateReminderUI()
	h.reminderWin.SetContent(ui)
	h.reminderWin.Show()
}

func (h *HydrapotionApp) CreateReminderUI() fyne.CanvasObject {
	consumed, goal, percent := h.GetProgress()

	// Titulo
	title := widget.NewLabelWithStyle("Bebe Agua!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	// Progreso actual
	progressText := widget.NewLabel(fmt.Sprintf("Hoy: %d / %d ml (%.0f%%)", consumed, goal, percent))
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.SetValue(percent)

	// Botones para agregar agua
	addWaterContainer := container.NewGridWithColumns(3,
		widget.NewButton("150ml", func() {
			h.AddWater(150)
			h.closeReminderWindow()
		}),
		widget.NewButton("250ml", func() {
			h.AddWater(250)
			h.closeReminderWindow()
		}),
		widget.NewButton("500ml", func() {
			h.AddWater(500)
			h.closeReminderWindow()
		}),
	)

	// Botones de posponer
	snoozeContainer := container.NewGridWithColumns(3,
		widget.NewButton("5 min", func() {
			h.snoozeReminder(5)
		}),
		widget.NewButton("15 min", func() {
			h.snoozeReminder(15)
		}),
		widget.NewButton("30 min", func() {
			h.snoezeReminder(30)
		}),
	)

	// Boton cerrar
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
		h.showReminderPopup()
	})
}

func (h *HydrapotionApp) snoezeReminder(minutes int) {
	h.snoozeReminder(minutes)
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

// --- Chart Widget ---

type WaterChart struct {
	widget.BaseWidget
	data []HistoryEntry
}

func NewWaterChart(data []HistoryEntry) *WaterChart {
	c := &WaterChart{data: data}
	c.ExtendBaseWidget(c)
	return c
}

func (c *WaterChart) CreateRenderer() fyne.WidgetRenderer {
	return &waterChartRenderer{chart: c}
}

type waterChartRenderer struct {
	chart *WaterChart
	bars  []*canvas.Rectangle
}

func (r *waterChartRenderer) Layout(size fyne.Size) {
	if len(r.chart.data) == 0 {
		return
	}

	maxMl := 0
	for _, d := range r.chart.data {
		if d.Ml > maxMl {
			maxMl = d.Ml
		}
	}
	if maxMl == 0 {
		maxMl = 1
	}

	barWidth := size.Width / float32(len(r.chart.data))
	padding := float32(5)
	maxHeight := size.Height - 40

	for i, bar := range r.bars {
		if i < len(r.chart.data) {
			ml := r.chart.data[i].Ml
			barHeight := float32(ml) / float32(maxMl) * maxHeight
			if barHeight < 2 {
				barHeight = 2
			}

			bar.Resize(fyne.NewSize(barWidth-padding*2, barHeight))
			bar.Move(fyne.NewPos(
				float32(i)*barWidth+padding,
				size.Height-barHeight-20,
			))
		}
	}
}

func (r *waterChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

func (r *waterChartRenderer) Refresh() {
	r.Layout(r.chart.Size())
	canvas.Refresh(r.chart)
}

func (r *waterChartRenderer) Objects() []fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, len(r.bars))
	for i, b := range r.bars {
		objs[i] = b
	}
	return objs
}

func (r *waterChartRenderer) Destroy() {}

// --- Main UI ---

func (h *HydrapotionApp) CreateMainUI() fyne.CanvasObject {
	// Header con titulo
	title := widget.NewLabelWithStyle("Hydrapotion", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	// Progreso
	progressLabel := widget.NewLabel("0 / 0 ml (0%)")
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100

	// Botones de agregar agua
	addWaterContainer := container.NewGridWithColumns(3,
		widget.NewButton("150ml", func() { h.AddWater(150) }),
		widget.NewButton("250ml", func() { h.AddWater(250) }),
		widget.NewButton("500ml", func() { h.AddWater(500) }),
	)

	// Selector de mood
	moodLabel := widget.NewLabel("Estado mental:")
	moodSelect := widget.NewSelect([]string{"Bien", "Neutral", "Bajo", "Tenso"}, func(s string) {
		var m Mood
		switch s {
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
	moodSelect.SetSelected(h.settings.CurrentMood.String())

	// Settings
	weightLabel := widget.NewLabel(fmt.Sprintf("Peso: %d kg", h.settings.Weight))
	weightEntry := widget.NewEntry()
	weightEntry.SetPlaceHolder("Peso en kg")
	weightEntry.SetText(fmt.Sprintf("%d", h.settings.Weight))

	saveWeightBtn := widget.NewButton("Guardar Peso", func() {
		var w int
		fmt.Sscanf(weightEntry.Text, "%d", &w)
		if w > 0 {
			h.SetWeight(w)
			weightLabel.SetText(fmt.Sprintf("Peso: %d kg", w))
		}
	})

	// Chart de historial
	chartLabel := widget.NewLabel("Progreso semanal:")
	chart := NewWaterChart(h.GetWeeklyData())

	// Update callback
	h.onUpdate = func() {
		consumed, goal, percent := h.GetProgress()
		progressLabel.SetText(fmt.Sprintf("%d / %d ml (%.0f%%)", consumed, goal, percent))
		progressBar.SetValue(percent)
		chart.data = h.GetWeeklyData()
		chart.Refresh()
	}

	// Initial update
	h.onUpdate()

	// Layout principal
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		progressLabel,
		progressBar,
		widget.NewSeparator(),
		widget.NewLabel("Agregar agua:"),
		addWaterContainer,
		widget.NewSeparator(),
		moodLabel,
		moodSelect,
		widget.NewSeparator(),
		weightLabel,
		weightEntry,
		saveWeightBtn,
		widget.NewSeparator(),
		chartLabel,
		chart,
	)

	return container.NewPadded(content)
}

func (h *HydrapotionApp) setupSystemTray() {
	if desk, ok := h.app.(desktop.App); ok {
		// Menu del system tray
		menu := fyne.NewMenu("Hydrapotion",
			fyne.NewMenuItem("Agregar 150ml", func() { h.AddWater(150) }),
			fyne.NewMenuItem("Agregar 250ml", func() { h.AddWater(250) }),
			fyne.NewMenuItem("Agregar 500ml", func() { h.AddWater(500) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Mostrar ventana", func() {
				if h.mainWindow != nil {
					h.mainWindow.Show()
					h.mainWindow.RequestFocus()
				}
			}),
			fyne.NewMenuItem("Salir", func() {
				h.app.Quit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}

func main() {
	a := app.NewWithID("hydrapotion")
	w := a.NewWindow("Hydrapotion")

	// Tema oscuro
	a.Settings().SetTheme(theme.DarkTheme())

	h := NewHydrapotionApp(a)
	h.mainWindow = w

	// Setup system tray
	h.setupSystemTray()

	// UI principal
	ui := h.CreateMainUI()
	w.SetContent(ui)
	w.Resize(fyne.NewSize(400, 550))

	// Cuando cierra la ventana principal, ocultar en vez de salir
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	// Iniciar timer de recordatorio
	h.startReminderTimer()

	w.ShowAndRun()
}
