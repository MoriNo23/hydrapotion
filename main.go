// Hydrapotion - Desktop hydration tracker with smart reminders
// Author: Mori
// License: MIT
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

// =============================================================================
// Data Models
// =============================================================================

// Mood represents user's mental/physical state.
// Affects daily water goal via bonus.
type Mood int

const (
	MoodWell    Mood = iota // Feeling good - no bonus
	MoodNeutral             // Neutral state
	MoodLow                 // Feeling low - +200ml bonus
	MoodTense               // Feeling tense/stressed - +400ml bonus
)

// String returns Spanish label for mood.
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

// Settings stores user preferences.
// Persisted to JSON in ~/.hydrapotion/
type Settings struct {
	Weight           int    `json:"weight"`
	TodayConsumed    int    `json:"today_consumed"`
	LastResetDate    string `json:"last_reset_date"`
	Language         string `json:"language"`
	Location         string `json:"location"`
	ReminderInterval int    `json:"reminder_interval"` // seconds
	CurrentMood      Mood   `json:"current_mood"`
}

// HistoryEntry represents one day's water intake.
type HistoryEntry struct {
	Date string `json:"date"` // YYYY-MM-DD format
	Ml   int    `json:"ml"`
}

// =============================================================================
// Application State
// =============================================================================

// HydrapotionApp manages all app state and UI.
type HydrapotionApp struct {
	settings    Settings       // User settings
	history     []HistoryEntry // Daily history
	mu          sync.RWMutex   // Thread-safe access
	dataDir     string         // ~/.hydrapotion path
	reminder    *time.Timer    // Reminder timer
	onUpdate    func()         // UI update callback
	mainWindow  fyne.Window    // Primary window
	reminderWin fyne.Window    // Reminder popup
	app         fyne.App       // Fyne app instance
}

// NewHydrapotionApp creates new app instance.
// Initializes data directory and loads persisted data.
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

// dataFile returns full path for a data file in dataDir.
func (h *HydrapotionApp) dataFile(name string) string {
	return filepath.Join(h.dataDir, name)
}

// =============================================================================
// Persistence - Load/Save
// =============================================================================

// loadSettings reads settings from JSON file.
// No error handling needed - defaults used if missing.
func (h *HydrapotionApp) loadSettings() {
	data, err := os.ReadFile(h.dataFile("settings.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, &h.settings)
}

// saveSettings writes settings to JSON file.
func (h *HydrapotionApp) saveSettings() {
	data, _ := json.MarshalIndent(h.settings, "", " ")
	os.WriteFile(h.dataFile("settings.json"), data, 0644)
}

// loadHistory reads history from JSON file.
func (h *HydrapotionApp) loadHistory() {
	data, err := os.ReadFile(h.dataFile("history.json"))
	if err != nil {
		return
	}
	json.Unmarshal(data, &h.history)
}

// saveHistory writes history to JSON file.
func (h *HydrapotionApp) saveHistory() {
	data, _ := json.MarshalIndent(h.history, "", " ")
	os.WriteFile(h.dataFile("history.json"), data, 0644)
}

// =============================================================================
// Core Business Logic
// =============================================================================

// todayStr returns current date as YYYY-MM-DD string.
func (h *HydrapotionApp) todayStr() string {
	return time.Now().Format("2006-01-02")
}

// resetIfNewDay resets daily consumption if date changed.
func (h *HydrapotionApp) resetIfNewDay() {
	today := h.todayStr()
	if h.settings.LastResetDate != today {
		h.settings.TodayConsumed = 0
		h.settings.LastResetDate = today
		h.settings.CurrentMood = MoodNeutral
		h.saveSettings()
	}
}

// CalculateGoal computes daily water goal.
// Base: 35ml per kg body weight.
// Mood bonus: +200ml (low) or +400ml (tense).
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

// AddWater records water intake.
// Updates settings, history, persists, triggers UI update.
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

// SetMood updates user's mood.
func (h *HydrapotionApp) SetMood(mood Mood) {
	h.mu.Lock()
	h.settings.CurrentMood = mood
	h.saveSettings()
	h.mu.Unlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}
}

// SetWeight updates user's weight.
func (h *HydrapotionApp) SetWeight(weight int) {
	h.mu.Lock()
	h.settings.Weight = weight
	h.saveSettings()
	h.mu.Unlock()

	if h.onUpdate != nil {
		h.onUpdate()
	}
}

// SetReminderInterval updates reminder frequency.
func (h *HydrapotionApp) SetReminderInterval(seconds int) {
	h.mu.Lock()
	h.settings.ReminderInterval = seconds
	h.saveSettings()
	h.mu.Unlock()
}

// GetProgress returns current hydration status.
// Thread-safe read with auto day reset.
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

// GetWeeklyData returns last 7 days of history.
// Used for weekly chart.
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

// =============================================================================
// Reminders
// =============================================================================

// startReminderTimer schedules next reminder popup.
func (h *HydrapotionApp) startReminderTimer() {
	if h.reminder != nil {
		h.reminder.Stop()
	}

	duration := time.Duration(h.settings.ReminderInterval) * time.Second
	h.reminder = time.AfterFunc(duration, func() {
		h.showReminderPopup()
	})
}

// showReminderPopup displays reminder notification and popup window.
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

// CreateReminderUI builds reminder popup content.
func (h *HydrapotionApp) CreateReminderUI() fyne.CanvasObject {
	consumed, goal, percent := h.GetProgress()

	title := widget.NewLabelWithStyle("Bebe Agua!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	progressText := widget.NewLabel(fmt.Sprintf("Hoy: %d / %d ml (%.0f%%)", consumed, goal, percent))
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100
	progressBar.SetValue(percent)

	// Water buttons with sound
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

	// Snooze buttons
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

// snoozeReminder delays reminder by specified minutes.
func (h *HydrapotionApp) snoozeReminder(minutes int) {
	h.closeReminderWindow()
	if h.reminder != nil {
		h.reminder.Stop()
	}
	h.reminder = time.AfterFunc(time.Duration(minutes)*time.Minute, func() {
		h.showReminderPopup()
	})
}

// closeReminderWindow hides reminder popup.
func (h *HydrapotionApp) closeReminderWindow() {
	if h.reminderWin != nil {
		h.reminderWin.Hide()
	}
}

// =============================================================================
// System Notifications
// =============================================================================

// notifySend shows Linux desktop notification.
func notifySend(title, message string) {
	cmd := exec.Command("notify-send", "-i", "water", title, message)
	cmd.Run()
}

// =============================================================================
// Sound Effects
// =============================================================================

// appSoundDir stores sounds in user data directory.
// Initialized on first playSound() call.
var appSoundDir string

// initSounds copies sound files from install dir to user data dir.
// Called lazily to ensure sounds work from any working directory.
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

// copyFile copies audio file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// playSound plays audio file using PulseAudio.
// Initializes sounds on first call.
func playSound(soundPath string) {
	if appSoundDir == "" {
		initSounds()
	}
	cmd := exec.Command("paplay", filepath.Join(appSoundDir, soundPath))
	cmd.Run()
}

// =============================================================================
// Chart Widget - Custom Fyne widget
// =============================================================================

// WaterChart displays weekly progress as bar chart.
// Custom Fyne widget rendering daily bars.
type WaterChart struct {
	widget.BaseWidget
	data []HistoryEntry
}

// NewWaterChart creates chart with given data.
func NewWaterChart(data []HistoryEntry) *WaterChart {
	c := &WaterChart{data: data}
	c.ExtendBaseWidget(c)
	return c
}

// CreateRenderer returns custom renderer.
func (c *WaterChart) CreateRenderer() fyne.WidgetRenderer {
	return &waterChartRenderer{chart: c}
}

// waterChartRenderer handles chart rendering.
type waterChartRenderer struct {
	chart *WaterChart
	bars  []*canvas.Rectangle
}

// Layout positions bar rectangles within chart area.
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

// MinSize returns minimum chart size.
func (r *waterChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100)
}

// Refresh re-renders chart.
func (r *waterChartRenderer) Refresh() {
	r.Layout(r.chart.Size())
	canvas.Refresh(r.chart)
}

// Objects returns bar rectangles.
func (r *waterChartRenderer) Objects() []fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, len(r.bars))
	for i, b := range r.bars {
		objs[i] = b
	}
	return objs
}

// Destroy is no-op for this renderer.
func (r *waterChartRenderer) Destroy() {}

// =============================================================================
// Main UI
// =============================================================================

// CreateMainUI builds primary window content.
func (h *HydrapotionApp) CreateMainUI() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Hydrapotion", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	progressLabel := widget.NewLabel("0 / 0 ml (0%)")
	progressBar := widget.NewProgressBar()
	progressBar.Min = 0
	progressBar.Max = 100

	// Water buttons with sound
	addWaterContainer := container.NewGridWithColumns(3,
		widget.NewButton("150ml", func() {
			h.AddWater(150)
			playSound("add_water.wav")
		}),
		widget.NewButton("250ml", func() {
			h.AddWater(250)
			playSound("add_water.wav")
		}),
		widget.NewButton("500ml", func() {
			h.AddWater(500)
			playSound("add_water.wav")
		}),
	)

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

	chartLabel := widget.NewLabel("Progreso semanal:")
	chart := NewWaterChart(h.GetWeeklyData())

	h.onUpdate = func() {
		consumed, goal, percent := h.GetProgress()
		progressLabel.SetText(fmt.Sprintf("%d / %d ml (%.0f%%)", consumed, goal, percent))
		progressBar.SetValue(percent)
		chart.data = h.GetWeeklyData()
		chart.Refresh()
	}

	h.onUpdate()

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

// =============================================================================
// System Tray
// =============================================================================

// setupSystemTray creates tray menu with quick actions.
func (h *HydrapotionApp) setupSystemTray() {
	if desk, ok := h.app.(desktop.App); ok {
		menu := fyne.NewMenu("Hydrapotion",
			fyne.NewMenuItem("Agregar 150ml", func() { h.AddWater(150) }),
			fyne.NewMenuItem("Agregar 250ml", func() { h.AddWater(250) }),
			fyne.NewMenuItem("Agregar 500ml", func() { h.AddWater(500) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Cambiar intervalo recordatorio", func() {
				h.showIntervalDialog()
			}),
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

// showIntervalDialog shows dialog to change reminder interval.
func (h *HydrapotionApp) showIntervalDialog() {
	dialog := h.app.NewWindow("Intervalo de Recordatorio")
	dialog.Resize(fyne.NewSize(300, 200))

	currentMin := h.settings.ReminderInterval / 60

	label := widget.NewLabel(fmt.Sprintf("Intervalo actual: %d minutos", currentMin))
	label.Alignment = fyne.TextAlignCenter

	options := []string{"5 min", "10 min", "15 min", "30 min", "45 min", "60 min"}
	selectWidget := widget.NewSelect(options, func(s string) {
		var minutes int
		fmt.Sscanf(s, "%d", &minutes)
		h.SetReminderInterval(minutes * 60)
		dialog.Close()
	})
	selectWidget.SetSelected(fmt.Sprintf("%d min", currentMin))

	infoLabel := widget.NewLabel("Selecciona el nuevo intervalo:")
	infoLabel.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		label,
		widget.NewSeparator(),
		infoLabel,
		selectWidget,
		layout.NewSpacer(),
		widget.NewButton("Cancelar", func() { dialog.Close() }),
	)

	dialog.SetContent(container.NewPadded(content))
	dialog.Show()
}

// =============================================================================
// Main Entry Point
// =============================================================================

func main() {
	a := app.NewWithID("hydrapotion")
	w := a.NewWindow("Hydrapotion")

	// Dark theme
	a.Settings().SetTheme(theme.DarkTheme())

	h := NewHydrapotionApp(a)
	h.mainWindow = w

	h.setupSystemTray()

	ui := h.CreateMainUI()
	w.SetContent(ui)
	w.Resize(fyne.NewSize(400, 550))

	// Hide instead of close - tray app behavior
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	h.startReminderTimer()

	w.ShowAndRun()
}
