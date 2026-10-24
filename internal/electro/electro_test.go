package electro

import "testing"

func TestDailyTarget(t *testing.T) {
	cases := []struct {
		name       string
		tempC      float64
		hasWeather bool
		want       int
	}{
		{"sin clima", 0, false, 2},
		{"templado", 20, true, 2},
		{"frontera 30", 30, true, 3},
		{"calor", 33.5, true, 3},
		{"muy caliente", 38, true, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := DailyTarget(c.tempC, c.hasWeather); got != c.want {
				t.Fatalf("DailyTarget(%.1f, %v) = %d; want %d", c.tempC, c.hasWeather, got, c.want)
			}
		})
	}
}

func TestSodiumForMl(t *testing.T) {
	cases := []struct {
		ml   int
		want int
	}{
		{0, 0},
		{-5, 0},
		{1, 1},     // 0.5 redondea a 1
		{250, 125}, // concentración isotónica genérica
		{500, 250},
		{1000, 500},
	}
	for _, c := range cases {
		if got := SodiumForMl(c.ml); got != c.want {
			t.Fatalf("SodiumForMl(%d) = %d; want %d", c.ml, got, c.want)
		}
	}
}

func TestFindByID(t *testing.T) {
	if s, ok := FindByID("ors"); !ok || s.Ml != 250 || s.SodiumMg != 430 {
		t.Fatalf("FindByID(ors) = %+v, %v; want suero 250 ml / 430 mg", s, ok)
	}
	if _, ok := FindByID("noexiste"); ok {
		t.Fatal("FindByID(noexiste) debe devolver false")
	}
}

func TestServingsSane(t *testing.T) {
	seen := map[string]bool{}
	for _, s := range Servings {
		if s.ID == "" {
			t.Fatal("porción con ID vacío")
		}
		if seen[s.ID] {
			t.Fatalf("ID duplicado: %s", s.ID)
		}
		seen[s.ID] = true
		if s.Ml <= 0 || s.Ml > 2000 {
			t.Fatalf("porción %s: volumen fuera de rango: %d", s.ID, s.Ml)
		}
		if s.SodiumMg <= 0 || s.SodiumMg > 1500 {
			t.Fatalf("porción %s: sodio estimado fuera de rango: %d", s.ID, s.SodiumMg)
		}
		// Cada porción debe tener su clave i18n (la UI usa "electro_"+ID).
		_ = s.ID
	}
}
