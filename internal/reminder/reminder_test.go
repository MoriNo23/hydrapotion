package reminder

import "testing"

func TestIntervalSeconds(t *testing.T) {
	cases := []struct {
		name string
		in   Input
		want int
	}{
		// Base 7200 s (2 h). Factores: mañana 0.8, mediodía 0.7, tarde 1.0, noche 2.0.
		{"mañana sin clima", Input{7200, 9, 0, 0, false}, 5760},
		{"mediodía sin clima", Input{7200, 14, 0, 0, false}, 5040},
		{"tarde sin clima", Input{7200, 20, 0, 0, false}, 7200},
		{"noche sin clima", Input{7200, 23, 0, 0, false}, 14400},
		{"madrugada sin clima", Input{7200, 3, 0, 0, false}, 14400},

		// Clima acelera el recordatorio (noche 2.0 * 0.5 = 1.0 → 7200).
		{"noche calor extremo", Input{7200, 23, 36, 50, true}, 7200},
		{"mediodía calor extremo (clamp factor 0.5)", Input{7200, 14, 36, 50, true}, 3600},
		{"tarde calor 30", Input{7200, 20, 30, 50, true}, 4320},

		// Humedad >70% multiplica 0.9 adicional.
		{"tarde calor 25 húmedo", Input{7200, 20, 25, 80, true}, 7200 * 0.8 * 0.9},

		// Clamps.
		{"clamp mínimo 30 min", Input{600, 20, 36, 50, true}, 1800},
		{"clamp máximo 6 h", Input{10800, 23, 0, 0, false}, 21600},
		{"base inválida usa 2 h", Input{0, 20, 0, 0, false}, 7200},
		{"base negativa usa 2 h", Input{-5, 9, 0, 0, false}, 5760},
	}
	for _, c := range cases {
		if got := IntervalSeconds(c.in); got != c.want {
			t.Errorf("%s: IntervalSeconds(%+v) = %d; want %d", c.name, c.in, got, c.want)
		}
	}
}

func TestIntervalMonotonicHotterIsSooner(t *testing.T) {
	// A más calor (misma hora y humedad), igual o menos espera.
	prev := IntervalSeconds(Input{7200, 20, 0, 40, true})
	for _, temp := range []float64{25, 30, 35, 40} {
		cur := IntervalSeconds(Input{7200, 20, temp, 40, true})
		if cur > prev {
			t.Errorf("con más calor el intervalo creció: %.0f° → %d > %d", temp, cur, prev)
		}
		prev = cur
	}
}
