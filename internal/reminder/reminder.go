// Package reminder — intervalo dinámico entre recordatorios.
// Puro y testeable: el intervalo base configurado se modula por hora del día
// y por clima, con clamp final [30 min, 6 h].
package reminder

import "time"

// Input — datos que alimentan el cálculo del intervalo.
type Input struct {
        BaseIntervalSeconds int // configurado por el usuario (default 7200 = 2 h)
        Hour                int // hora local 0-23
        TempC               float64
        HumidityPct         float64
        HasWeather          bool
}

// IntervalSeconds — segundos que deben pasar entre notificaciones.
func IntervalSeconds(in Input) int {
        base := in.BaseIntervalSeconds
        if base <= 0 {
                base = 7200
        }

        var timeFactor float64
        switch {
        case in.Hour >= 6 && in.Hour < 12:
                timeFactor = 0.8
        case in.Hour >= 12 && in.Hour < 18:
                timeFactor = 0.7
        case in.Hour >= 18 && in.Hour < 22:
                timeFactor = 1.0
        default:
                timeFactor = 2.0
        }

        weatherFactor := 1.0
        if in.HasWeather {
                switch {
                case in.TempC >= 35:
                        weatherFactor = 0.5
                case in.TempC >= 30:
                        weatherFactor = 0.6
                case in.TempC >= 25:
                        weatherFactor = 0.8
                }
                if in.HumidityPct > 70 {
                        weatherFactor *= 0.9
                }
        }

        factor := timeFactor * weatherFactor
        if factor < 0.5 {
                factor = 0.5
        }

        d := time.Duration(float64(base) * factor * float64(time.Second))
        if d < 30*time.Minute {
                d = 30 * time.Minute
        }
        if d > 6*time.Hour {
                d = 6 * time.Hour
        }
        return int(d.Seconds())
}
