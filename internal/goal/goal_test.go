package goal

import "testing"

func TestBase(t *testing.T) {
	cases := []struct {
		weight, want int
	}{
		{70, 2450},
		{50, 1750},
		{100, 3500},
		{0, 0},
		{-5, 0},
	}
	for _, c := range cases {
		if got := Base(c.weight); got != c.want {
			t.Errorf("Base(%d) = %d; want %d", c.weight, got, c.want)
		}
	}
}

func TestHeightBonus(t *testing.T) {
	cases := []struct {
		height, want int
	}{
		{150, 0},
		{149, 0},
		{151, 5},
		{180, 150},
		{200, 250},
	}
	for _, c := range cases {
		if got := HeightBonus(c.height); got != c.want {
			t.Errorf("HeightBonus(%d) = %d; want %d", c.height, got, c.want)
		}
	}
}

func TestMoodBonus(t *testing.T) {
	cases := []struct {
		mood string
		want int
	}{
		{"Tenso", 400},
		{"Bajo", 200},
		{"Bien", 0},
		{"Neutral", 0},
		{"", 0},
		{"otro", 0},
	}
	for _, c := range cases {
		if got := MoodBonus(c.mood); got != c.want {
			t.Errorf("MoodBonus(%q) = %d; want %d", c.mood, got, c.want)
		}
	}
}

func TestWeatherBonus(t *testing.T) {
	cases := []struct {
		name           string
		temp, humidity float64
		hasWeather     bool
		want           int
	}{
		{"sin clima", 30, 50, false, 0},
		{"frío seco", 10, 40, true, 0},
		{"umbral 25 exacto", 25, 40, true, 200},
		{"frontera 24.9", 24.9, 40, true, 0},
		{"umbral 30 exacto", 30, 40, true, 400},
		{"umbral 35 exacto", 35, 40, true, 600},
		{"húmedo 60 exacto", 20, 60, true, 0},
		{"húmedo 61", 20, 61, true, 100},
		{"calor + húmedo", 32, 80, true, 500},
		{"extremo + húmedo", 36, 70, true, 700},
	}
	for _, c := range cases {
		if got := WeatherBonus(c.temp, c.humidity, c.hasWeather); got != c.want {
			t.Errorf("%s: WeatherBonus(%.1f, %.0f, %v) = %d; want %d",
				c.name, c.temp, c.humidity, c.hasWeather, got, c.want)
		}
	}
}

func TestTotal(t *testing.T) {
	// 70kg → 2450 + 180cm → 150 + Tenso → 400 + 32°/80% → 500 = 3500
	got := Total(70, 180, "Tenso", 32, 80, true)
	if got != 3500 {
		t.Errorf("Total() = %d; want 3500", got)
	}
	// Sin clima y sin extras: solo base.
	if got := Total(70, 150, "Neutral", 0, 0, false); got != 2450 {
		t.Errorf("Total() = %d; want 2450", got)
	}
}
