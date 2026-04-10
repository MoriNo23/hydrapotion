package main

import (
	"context"
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icon.png
var iconData []byte

// Mood representa el estado emocional
type Mood string

const (
	MoodWell   Mood = "well"   // Bien - contento, tranquilo, feliz
	MoodNeutral Mood = "neutral" // Neutral - ni bien ni mal, normal
	MoodLow    Mood = "low"    // Bajo - triste, deprimido, melancolico
	MoodTense  Mood = "tense"  // Tenso - estresado, ansioso, frustrado
)

// Settings estructura principal
type Settings struct {
	Weight           int    `json:"weight"`
	TodayConsumed    int    `json:"today_consumed"`
	DailyGoal        int    `json:"daily_goal"`
	Language         string `json:"language"`
	Location         string `json:"location"`
	LastResetDate    string `json:"last_reset_date"`
	ReminderInterval int    `json:"reminder_interval"`
	CurrentMood      Mood   `json:"current_mood"`
}

// HistoryDay entrada de historial
type HistoryDay struct {
	Day  string `json:"day"`
	Ml   int    `json:"ml"`
	Date string `json:"date"`
}

// MoodEntry registro de mood
type MoodEntry struct {
	Date string `json:"date"`
	Mood Mood   `json:"mood"`
	Ml   int    `json:"ml"` // agua consumida en ese momento
}

// App estado de la aplicacion
type App struct {
	settings    Settings
	history     map[string]int
	moodHistory []MoodEntry
	mu          sync.RWMutex
	ctx         context.Context
	shouldClose bool
}

var appInstance *App

func main() {
	appInstance = &App{
		settings: Settings{
			Weight:           70,
			TodayConsumed:    0,
			DailyGoal:        2450,
			Language:         "es",
			Location:         "",
			LastResetDate:    todayStr(),
			ReminderInterval: 1800,
			CurrentMood:      MoodNeutral,
		},
		history:     make(map[string]int),
		moodHistory: make([]MoodEntry, 0),
		shouldClose: false,
	}
	appInstance.loadSettings()
	appInstance.loadHistory()
	appInstance.loadMoodHistory()

	// Iniciar systray en goroutine
	go systray.Run(appInstance.onReady, appInstance.onExit)

	err := wails.Run(&options.App{
		Title: "Hydrapotion",
		Width:            340,
		Height:           760,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 11, G: 23, B: 32, A: 1},
		OnStartup:        appInstance.startup,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			if appInstance.shouldClose {
				return false
			}
			runtime.WindowHide(ctx)
			return true
		},
		Bind: []interface{}{
			appInstance,
		},
		Linux: &linux.Options{
			ProgramName: "Hydrapotion",
			WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func (a *App) onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Hydrapotion")
	systray.SetTooltip("Hydrapotion - Tracker de hidratacion")

	mShow := systray.AddMenuItem("Mostrar", "Mostrar ventana")
	systray.AddSeparator()
	mAdd150 := systray.AddMenuItem("+150 ml", "Agregar 150ml")
	mAdd250 := systray.AddMenuItem("+250 ml", "Agregar 250ml")
	mAdd500 := systray.AddMenuItem("+500 ml", "Agregar 500ml")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Salir", "Salir de la aplicacion")

	mShow.Click(func() {
		if a.ctx != nil {
			runtime.WindowShow(a.ctx)
			runtime.WindowSetAlwaysOnTop(a.ctx, true)
			runtime.WindowSetAlwaysOnTop(a.ctx, false)
		}
	})

	mAdd150.Click(func() {
		appInstance.AddWater(150)
	})

	mAdd250.Click(func() {
		appInstance.AddWater(250)
	})

	mAdd500.Click(func() {
		appInstance.AddWater(500)
	})

	mQuit.Click(func() {
		a.shouldClose = true
		systray.Quit()
		runtime.Quit(a.ctx)
	})
}

func (a *App) onExit() {
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

	return a.settings
}

func (a *App) SetMood(mood Mood) Settings {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.settings.CurrentMood = mood
	a.saveSettings()

	// Guardar en historial de moods
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

// GetMoodRecommendation devuelve recomendacion segun mood y temperatura
func (a *App) GetMoodRecommendation(temp int) map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	baseGoal := a.settings.DailyGoal
	mood := a.settings.CurrentMood

	// Ajustes segun mood
	multiplier := 1.0
	var recommendation string
	var adjustment string

	switch mood {
	case MoodWell:
		multiplier = 1.0
		recommendation = "¡Genial! Sigue asi"
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

	// Ajuste por temperatura
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
	a.settings.DailyGoal = weight * 35
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
	path := filepath.Join(getConfigDir(), "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &a.settings)
}

func (a *App) loadHistory() {
	path := filepath.Join(getConfigDir(), "history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &a.history)
}

func (a *App) loadMoodHistory() {
	path := filepath.Join(getConfigDir(), "mood_history.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &a.moodHistory)
}
