package main

import (
 "embed"
 "encoding/json"
 "log"
 "os"
 "path/filepath"
 "sync"
 "time"

 "github.com/wailsapp/wails/v3/pkg/application"
 "github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icon.png
var iconData []byte

// Mood representa el estado emocional
type Mood string

const (
	MoodWell    Mood = "well"
	MoodNeutral Mood = "neutral"
	MoodLow     Mood = "low"
	MoodTense   Mood = "tense"
)

// Settings estructura principal
type Settings struct {
	Weight          int    `json:"weight"`
	Height          int    `json:"height"`           // Estatura en cm
	TodayConsumed   int    `json:"today_consumed"`
	DailyGoal       int    `json:"daily_goal"`
	Language        string `json:"language"`
	Location        string `json:"location"`
	LastResetDate   string `json:"last_reset_date"`
	ReminderInterval int   `json:"reminder_interval"`
	CurrentMood      Mood  `json:"current_mood"`
}

// HistoryDay entrada de historial
type HistoryDay struct {
	Day   string `json:"day"`
	Ml    int    `json:"ml"`
	Date  string `json:"date"`
}

// MoodEntry registro de mood
type MoodEntry struct {
	Date string `json:"date"`
	Mood Mood   `json:"mood"`
	Ml   int    `json:"ml"`
}

// App servicio principal de la aplicacion
type App struct {
	settings      Settings
	history       map[string]int
	moodHistory   []MoodEntry
	mu            sync.RWMutex
	reminderTimer *time.Timer
	lastIntakeTime time.Time
	window        *application.WebviewWindow
	systray       *application.SystemTray
}

func main() {
	// Crear aplicacion
	app := application.New(application.Options{
		Name:        "Hydrapotion",
		Description: "Desktop hydration tracker",
		Services: []application.Service{
			application.NewService(NewApp()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

 // Crear ventana principal
 window := app.Window.NewWithOptions(application.WebviewWindowOptions{
 Title:            "Hydrapotion",
 Width:            340,
 Height:           760,
 BackgroundColour: application.NewRGB(11, 23, 32),
 URL:              "/",
 Hidden:           false,
 })

 // Maximizar la ventana al iniciar
 window.Maximise()

 // Ocultar en lugar de cerrar usando RegisterHook
 window.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
 e.Cancel()
 window.Hide()
 })

 // Guardar referencia a la ventana
 appService := NewApp()
 appService.window = window

 // Crear system tray
 systray := app.SystemTray.New()
 systray.SetIcon(iconData)
 systray.SetLabel("Hydrapotion")
 systray.SetTooltip("Hydrapotion - Tracker de hidratacion")

 // Menu del system tray (solo para click derecho)
 menu := app.NewMenu()
 menu.Add("Mostrar").OnClick(func(ctx *application.Context) {
 window.Show()
 window.Restore()
 window.Focus()
 })
 menu.AddSeparator()
 menu.Add("+150 ml").OnClick(func(ctx *application.Context) {
 appService.AddWater(150)
 })
 menu.Add("+250 ml").OnClick(func(ctx *application.Context) {
 appService.AddWater(250)
 })
 menu.Add("+500 ml").OnClick(func(ctx *application.Context) {
 appService.AddWater(500)
 })
 menu.AddSeparator()
 menu.Add("Salir").OnClick(func(ctx *application.Context) {
 app.Quit()
 })
 systray.SetMenu(menu)

 // Click izquierdo SOLO muestra la ventana (no toggle)
 systray.OnClick(func() {
 window.Show()
 window.Restore()
 window.Focus()
 window.SetAlwaysOnTop(true)
 window.SetAlwaysOnTop(false)
 })

 // Click derecho SOLO abre el menu (no muestra la ventana)
 systray.OnRightClick(func() {
 // El menu se abre automaticamente por SetMenu
 // No hacemos nada mas para evitar mostrar la ventana
 })

	appService.systray = systray

	// Registrar eventos
	application.RegisterEvent[map[string]interface{}]("show-reminder")

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// NewApp crea una nueva instancia del servicio App
func NewApp() *App {
	a := &App{
		settings: Settings{
			Weight:           70,
			Height:           170,
			TodayConsumed:    0,
			DailyGoal:        2450,
			Language:         "es",
			Location:         "",
			LastResetDate:    todayStr(),
			ReminderInterval: 1800,
			CurrentMood:      MoodNeutral,
		},
		history:       make(map[string]int),
		moodHistory:   make([]MoodEntry, 0),
		lastIntakeTime: time.Now(),
	}
	a.loadSettings()
	a.loadHistory()
	a.loadMoodHistory()
	return a
}

// calculateDailyGoal calcula el objetivo diario basado en peso Y estatura
func (a *App) calculateDailyGoal() int {
	weight := a.settings.Weight
	height := a.settings.Height

	// Formula: 35ml por kg + ajuste por estatura
	// Por cada 10cm arriba de 150cm, agregar 100ml extra
	baseGoal := weight * 35

	heightAdjustment := 0
	if height > 150 {
		heightAdjustment = ((height - 150) / 10) * 100
	}

	return baseGoal + heightAdjustment
}

// Startup llamado cuando la app inicia
func (a *App) Startup() {
	go a.startReminderTimer()
}

// --- Reminder System ---

func (a *App) startReminderTimer() {
	a.mu.RLock()
	interval := time.Duration(a.settings.ReminderInterval) * time.Second
	a.mu.RUnlock()

	a.mu.Lock()
	a.lastIntakeTime = time.Now()
	a.mu.Unlock()

	a.mu.Lock()
	a.reminderTimer = time.AfterFunc(interval, a.showReminder)
	a.mu.Unlock()
}

func (a *App) resetReminderTimer() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.lastIntakeTime = time.Now()

	if a.reminderTimer != nil {
		a.reminderTimer.Stop()
	}

	interval := time.Duration(a.settings.ReminderInterval) * time.Second
	a.reminderTimer = time.AfterFunc(interval, a.showReminder)
}

func (a *App) showReminder() {
 // Emitir evento al frontend
 if a.window != nil {
 a.window.EmitEvent("show-reminder", map[string]interface{}{
 "consumed": a.settings.TodayConsumed,
 "goal":     a.settings.DailyGoal,
 })
 a.window.Show()
 }
}

// SnoozeReminder pospone el recordatorio
func (a *App) SnoozeReminder(minutes int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.reminderTimer != nil {
		a.reminderTimer.Stop()
	}

	interval := time.Duration(minutes) * time.Minute
	a.reminderTimer = time.AfterFunc(interval, a.showReminder)
}

// DismissReminder reinicia el timer
func (a *App) DismissReminder() {
	a.resetReminderTimer()
}

// --- API Methods ---

func (a *App) GetSettings() Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	today := todayStr()
	if a.settings.LastResetDate != today {
		a.settings.TodayConsumed = 0
		a.settings.LastResetDate = today
		a.settings.CurrentMood = MoodNeutral
		a.saveSettings()
	}

	return a.settings
}

func (a *App) AddWater(ml int) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	today := todayStr()
	if a.settings.LastResetDate != today {
		a.settings.TodayConsumed = 0
		a.settings.LastResetDate = today
	}

	a.settings.TodayConsumed += ml
	a.history[today] += ml

	a.saveSettings()
	a.saveHistory()

	// Reiniciar timer
	a.lastIntakeTime = time.Now()
	if a.reminderTimer != nil {
		a.reminderTimer.Stop()
		interval := time.Duration(a.settings.ReminderInterval) * time.Second
		a.reminderTimer = time.AfterFunc(interval, a.showReminder)
	}

	return a.settings
}

func (a *App) SetMood(mood Mood) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.CurrentMood = mood
	a.saveSettings()

	entry := MoodEntry{
		Date: time.Now().Format("2006-01-02 15:04:05"),
		Mood: mood,
		Ml:   a.settings.TodayConsumed,
	}
	a.moodHistory = append(a.moodHistory, entry)
	a.saveMoodHistory()

	return a.settings
}

func (a *App) GetMoodHistory() []MoodEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.moodHistory
}

func (a *App) GetMoodRecommendation(temp int) map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	baseGoal := a.settings.DailyGoal
	mood := a.settings.CurrentMood

	multiplier := 1.0
	var recommendation string
	var adjustment string

	switch mood {
	case MoodWell:
		multiplier = 1.0
		recommendation = "Genial! Sigue asi"
		adjustment = "Normal"
	case MoodNeutral:
		multiplier = 1.0
		recommendation = "Mantente hidratado"
		adjustment = "Normal"
	case MoodLow:
		multiplier = 1.1
		recommendation = "El agua ayuda a mejorar el animo"
		adjustment = "+10% recomendado"
	case MoodTense:
		multiplier = 1.15
		recommendation = "El estres deshidrata, bebe mas"
		adjustment = "+15% recomendado"
	}

	if temp >= 30 {
		multiplier += 0.1
		adjustment = "Extra por calor"
	}

	adjustedGoal := int(float64(baseGoal) * multiplier)

	return map[string]interface{}{
		"mood":          string(mood),
		"base_goal":     baseGoal,
		"adjusted_goal": adjustedGoal,
		"recommendation": recommendation,
		"adjustment":    adjustment,
		"multiplier":    multiplier,
	}
}

func (a *App) SetWeight(weight int) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.Weight = weight
	a.settings.DailyGoal = a.calculateDailyGoal()
	a.saveSettings()

	return a.settings
}

// SetHeight establece la estatura y recalcula el objetivo
func (a *App) SetHeight(height int) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.Height = height
	a.settings.DailyGoal = a.calculateDailyGoal()
	a.saveSettings()

	return a.settings
}

func (a *App) SetLanguage(language string) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.Language = language
	a.saveSettings()

	return a.settings
}

func (a *App) SetLocation(location string) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.Location = location
	a.saveSettings()

	return a.settings
}

func (a *App) SetReminderInterval(interval int) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.ReminderInterval = interval
	a.saveSettings()

	return a.settings
}

func (a *App) GetWeeklyData() []HistoryDay {
	return a.getHistoryData(7, true)
}

func (a *App) GetMonthlyData() []HistoryDay {
	return a.getHistoryData(30, false)
}

func (a *App) getHistoryData(days int, useWeekday bool) []HistoryDay {
	a.mu.RLock()
	defer a.mu.RUnlock()

	result := make([]HistoryDay, 0, days)
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")

		var dayStr string
		if useWeekday {
			dayStr = getWeekdayShort(date.Weekday())
		} else {
			dayStr = date.Format("02")
		}

		ml := a.history[dateStr]

		result = append(result, HistoryDay{
			Day:  dayStr,
			Ml:   ml,
			Date: dateStr,
		})
	}

	return result
}

func getWeekdayShort(d time.Weekday) string {
	switch d {
	case time.Monday:
		return "L"
	case time.Tuesday:
		return "M"
	case time.Wednesday:
		return "X"
	case time.Thursday:
		return "J"
	case time.Friday:
		return "V"
	case time.Saturday:
		return "S"
	case time.Sunday:
		return "D"
	}
	return "?"
}

func todayStr() string {
	return time.Now().Format("2006-01-02")
}

// --- Persistence ---

func getConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "hydrapotion")
}

func (a *App) saveSettings() {
	dir := getConfigDir()
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, "settings.json")
	data, _ := json.MarshalIndent(a.settings, "", " ")
	os.WriteFile(path, data, 0644)
}

func (a *App) saveHistory() {
	dir := getConfigDir()
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, "history.json")
	data, _ := json.MarshalIndent(a.history, "", " ")
	os.WriteFile(path, data, 0644)
}

func (a *App) saveMoodHistory() {
	dir := getConfigDir()
	os.MkdirAll(dir, 0755)

	path := filepath.Join(dir, "mood_history.json")
	data, _ := json.MarshalIndent(a.moodHistory, "", " ")
	os.WriteFile(path, data, 0644)
}

func (a *App) loadSettings() {
	dir := getConfigDir()
	path := filepath.Join(dir, "settings.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	json.Unmarshal(data, &a.settings)
}

func (a *App) loadHistory() {
	dir := getConfigDir()
	path := filepath.Join(dir, "history.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	json.Unmarshal(data, &a.history)
}

func (a *App) loadMoodHistory() {
	dir := getConfigDir()
	path := filepath.Join(dir, "mood_history.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	json.Unmarshal(data, &a.moodHistory)
}
