# Hydrapotion

App de seguimiento de hidratacion construida con Go + Wails.

## Caracteristicas

- Tracker de agua diario
- Historial de consumo
- Clima integrado (Open-Meteo)
- Sistema de mood que afecta recomendaciones
- Sonido ASMR de teclado

## Requisitos

- Go 1.24+
- Node.js 18+
- Wails CLI v2.12+

## Desarrollo

```bash
# Instalar Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Ejecutar en modo desarrollo
wails dev

# Compilar
wails build
```

## Compilacion

El binario se genera en `build/bin/hydrapotion`

Tiempo de compilacion: ~8 segundos

## Licencia

MIT
