package main

// headless.go — modo sin GUI para hydrapotion.
//
// Permite delegar los recordatorios a systemd timers (Linux) y registrar
// agua desde la terminal, sin iniciar Fyne/OpenGL. El proceso --remind
// vive menos de 1 segundo y consume ~10 MB, frente a los ~200 MB de la GUI.
//
// Comandos:
//
//      hydrapotion --remind      → evalúa si toca notificar y sale (lo llama el systemd timer)
//      hydrapotion add <ml>      → registra agua sin abrir la ventana
//      hydrapotion status        → progreso del día en texto plano
//      hydrapotion --version     → versión
//
// Estado compartido entre GUI y modo headless: ~/.hydrapotion/state.json
// (last_notified evita notificaciones duplicadas; last_weather_fetch evita
// martillar Open-Meteo). Cache de clima: ~/.hydrapotion/weather.json.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"
)

var version = "dev"

// ----------------------------------------------------------------
// Estado compartido (state.json)
// ----------------------------------------------------------------

type sharedState struct {
	LastNotified     int64 `json:"last_notified"`
	LastWeatherFetch int64 `json:"last_weather_fetch"`
}

func loadSharedState(path string) sharedState {
	var s sharedState
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &s)
	}
	return s
}

func (h *HydrapotionApp) saveSharedState() {
	data, _ := json.MarshalIndent(h.state, "", " ")
	os.WriteFile(h.dataFile("state.json"), data, 0644)
}

func (h *HydrapotionApp) markNotified(t time.Time) {
	h.mu.Lock()
	h.state.LastNotified = t.Unix()
	h.mu.Unlock()
	h.saveSharedState()
}

// ----------------------------------------------------------------
// Cache de clima (weather.json)
// ----------------------------------------------------------------

type weatherCache struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	FetchedAt   int64   `json:"fetched_at"`
}

func (h *HydrapotionApp) loadWeatherCache() {
	data, err := os.ReadFile(h.dataFile("weather.json"))
	if err != nil {
		return
	}
	var c weatherCache
	if json.Unmarshal(data, &c) == nil {
		h.weather = &WeatherData{Temperature: c.Temperature, Humidity: c.Humidity}
		h.state.LastWeatherFetch = c.FetchedAt
	}
}

func (h *HydrapotionApp) saveWeatherCache() {
	if h.weather == nil {
		return
	}
	c := weatherCache{
		Temperature: h.weather.Temperature,
		Humidity:    h.weather.Humidity,
		FetchedAt:   h.state.LastWeatherFetch,
	}
	data, _ := json.MarshalIndent(c, "", " ")
	os.WriteFile(h.dataFile("weather.json"), data, 0644)
}

// refreshWeatherIfStale re-descarga el clima si el cache tiene más de 1 hora.
// Best-effort: si falla la red se usa el cache y el recordatorio sigue vivo.
func (h *HydrapotionApp) refreshWeatherIfStale() {
	if h.settings.LocationLat == 0 && h.settings.LocationLon == 0 {
		return
	}
	if time.Since(time.Unix(h.state.LastWeatherFetch, 0)) < time.Hour {
		return
	}
	h.fetchWeather() // headless-safe: onUpdate es nil, no toca Fyne
	h.mu.Lock()
	h.state.LastWeatherFetch = time.Now().Unix()
	h.mu.Unlock()
	h.saveWeatherCache()
}

// ----------------------------------------------------------------
// Constructor headless (cero Fyne)
// ----------------------------------------------------------------

func newHeadlessApp() *HydrapotionApp {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".hydrapotion")
	os.MkdirAll(dataDir, 0755)

	h := &HydrapotionApp{
		dataDir:  dataDir,
		headless: true,
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
	h.state = loadSharedState(h.dataFile("state.json"))
	h.loadWeatherCache()
	return h
}

// ----------------------------------------------------------------
// Dispatch de subcomandos headless
// ----------------------------------------------------------------

// runHeadless procesa os.Args[1:] si corresponde a un subcomando sin GUI.
// Devuelve true si el argumento fue manejado (el caller debe retornar).
func runHeadless(args []string) bool {
	switch args[0] {
	case "--remind":
		runReminder()
		return true
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "uso: hydrapotion add <ml>   (ej: hydrapotion add 250)")
			os.Exit(2)
		}
		ml, err := strconv.Atoi(args[1])
		if err != nil || ml <= 0 || ml > 10000 {
			fmt.Fprintln(os.Stderr, "cantidad inválida:", args[1])
			os.Exit(2)
		}
		runAdd(ml)
		return true
	case "status":
		runStatus()
		return true
	case "--version", "version":
		fmt.Printf("hydrapotion %s\n", version)
		return true
	}
	// Flag no reconocido acá puede ser de la GUI; no manejar.
	return false
}

// ----------------------------------------------------------------
// runReminder — invocado por el systemd timer cada 15 min
// ----------------------------------------------------------------

func runReminder() {
	debug.SetGCPercent(50)
	debug.SetMemoryLimit(48 << 20)

	h := newHeadlessApp()
	h.refreshWeatherIfStale()

	c, g, p, _ := h.GetProgress()

	// Meta cumplida: no molestar.
	if g > 0 && c >= g {
		return
	}

	interval := h.dynamicInterval()
	last := time.Unix(h.state.LastNotified, 0)
	if time.Since(last) < interval {
		return
	}

	notifySend(
		T("notify_title"),
		fmt.Sprintf("%s · %d%% (%d/%d ml)", T("notify_body"), int(math.Round(p)), c, g),
	)
	h.markNotified(time.Now())
}

// ----------------------------------------------------------------
// runAdd / runStatus — CLI para registrar y consultar
// ----------------------------------------------------------------

func runAdd(ml int) {
	h := newHeadlessApp()
	h.AddWater(ml)
	c, g, p, _ := h.GetProgress()
	fmt.Printf("+%d ml → %d/%d ml (%.0f%%)\n", ml, c, g, p)
}

func runStatus() {
	h := newHeadlessApp()
	c, g, p, w := h.GetProgress()
	fmt.Printf("hidratación: %d/%d ml (%.0f%%)\n", c, g, p)
	if w != nil {
		fmt.Printf("clima: %.1f°C, %d%% humedad\n", w.Temperature, int(w.Humidity))
	}
	if h.state.LastNotified != 0 {
		fmt.Printf("último recordatorio: %s\n", time.Unix(h.state.LastNotified, 0).Format("15:04"))
	}
}
