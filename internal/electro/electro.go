// Package electro — lógica pura de electrolitos: porciones típicas, sodio
// estimado y objetivo diario. Sin estado ni dependencias de Fyne: testeable
// y compartible entre la GUI (main.go) y el modo headless (headless.go).
//
// Todo el sodio es ESTIMACIÓN informativa a partir de concentraciones
// habituales; no es consejo médico (ver DISCLAIMER del README).
package electro

// Serving — una porción típica de bebida o alimento recuperador de
// electrolitos, con el sodio que aporta de forma aproximada.
type Serving struct {
	ID       string // identificador interno (sufijo de clave i18n y de icono)
	Ml       int    // volumen de la porción
	SodiumMg int    // sodio estimado por porción (mg)
}

// Servings — porciones que ofrece la UI, con su sodio estimado:
//
//   - bebida isotónica 500 ml ≈ 250 mg Na (típico 400–500 mg Na/L)
//   - suero oral / sales de rehidratación 250 ml ≈ 430 mg Na
//     (fórmula OMS ~75 mmol/L ≈ 1,7 g Na/L)
//   - agua de coco 250 ml ≈ 105 mg Na (~400 mg/L; además aporta potasio)
//   - caldo / consomé 250 ml ≈ 500 mg Na (casero ligero; enorme variabilidad)
var Servings = []Serving{
	{"isotonic", 500, 250},
	{"ors", 250, 430},
	{"coconut", 250, 105},
	{"broth", 250, 500},
}

// NaPerMl — concentración genérica (isotónica) para estimar el sodio de
// volúmenes arbitrarios registrados por CLI (hydrapotion electro <ml>).
const NaPerMl = 0.5

// SodiumForMl estima el sodio de un volumen arbitrario. Volúmenes <= 0 → 0.
func SodiumForMl(ml int) int {
	if ml <= 0 {
		return 0
	}
	return int(float64(ml)*NaPerMl + 0.5)
}

// FindByID devuelve la porción con ese identificador.
func FindByID(id string) (Serving, bool) {
	for _, s := range Servings {
		if s.ID == id {
			return s, true
		}
	}
	return Serving{}, false
}

// DailyTarget — dosis de electrolitos de referencia por día:
// 2 de base; +1 si hace mucho calor (≥ 30 °C) por el sudor extra.
// Igual que el resto de fórmulas del programa, es una heurística descriptiva.
func DailyTarget(tempC float64, hasWeather bool) int {
	t := 2
	if hasWeather && tempC >= 30 {
		t++
	}
	return t
}
