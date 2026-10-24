// Package goal — cálculo de la meta diaria de hidratación.
// Lógica pura, sin estado ni dependencias de Fyne: testeable y compartible
// entre la GUI (main.go) y el modo headless (headless.go).
package goal

// MlPerKg — base de la meta: 35 ml por kilo de peso corporal.
const MlPerKg = 35

// Base — meta base por peso. Peso inválido (<=0) → 0 (la UI lo valida).
func Base(weightKg int) int {
	if weightKg <= 0 {
		return 0
	}
	return weightKg * MlPerKg
}

// HeightBonus — +5 ml por cada cm por encima de 150.
func HeightBonus(heightCm int) int {
	if heightCm <= 150 {
		return 0
	}
	return (heightCm - 150) * 5
}

// MoodBonus — ajuste por estado mental del día.
// Canonical: "Tenso" +400, "Bajo" +200, resto 0.
func MoodBonus(mood string) int {
	switch mood {
	case "Tenso":
		return 400
	case "Bajo":
		return 200
	}
	return 0
}

// WeatherBonus — ajuste por clima actual.
// Temperatura: >=35° +600, >=30° +400, >=25° +200. Humedad >60%: +100.
func WeatherBonus(tempC, humidityPct float64, hasWeather bool) int {
	if !hasWeather {
		return 0
	}
	b := 0
	switch {
	case tempC >= 35:
		b = 600
	case tempC >= 30:
		b = 400
	case tempC >= 25:
		b = 200
	}
	if humidityPct > 60 {
		b += 100
	}
	return b
}

// Total — meta diaria completa: base + altura + ánimo + clima.
func Total(weightKg, heightCm int, mood string, tempC, humidityPct float64, hasWeather bool) int {
	return Base(weightKg) +
		HeightBonus(heightCm) +
		MoodBonus(mood) +
		WeatherBonus(tempC, humidityPct, hasWeather)
}
