package main

// theme.go — tema "informe hidrológico" (mundo visual USGS).
// Papel hueso, tinta férrea, azul-hidro para datos, vermellón para sellos.
// La estrategia de color es Restrained: neutros tintados del mundo + un acento.

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Paleta del informe (compartida por widgets raster: tarjetas, hidrógrafo, resumen).
var (
	repPaper     = color.RGBA{0xF5, 0xF1, 0xE6, 0xFF} // papel hueso (fondo)
	repPanel     = color.RGBA{0xEC, 0xE7, 0xD8, 0xFF} // papel de panel
	repInk       = color.RGBA{0x22, 0x31, 0x3A, 0xFF} // tinta férrea
	repInkMuted  = color.RGBA{0x5C, 0x6A, 0x64, 0xFF} // tinta desvanecida (secundario)
	repHydro     = color.RGBA{0x1F, 0x5F, 0x7A, 0xFF} // azul-hidro (dato de agua)
	repHydroSoft = color.RGBA{0x1F, 0x5F, 0x7A, 0x26} // lavado hidro 15% (área de curva)
	repWash      = color.RGBA{0xE0, 0xE8, 0xEA, 0xFF} // lavado de hover sobre papel
	repVermilion = color.RGBA{0xA6, 0x3E, 0x28, 0xFF} // vermellón (sellos, meta superada; ≥4.5:1 sobre panel)
	repMoss      = color.RGBA{0x4A, 0x7C, 0x59, 0xFF} // verde musgo (día cumplido)
	repAmber     = color.RGBA{0x7E, 0x5A, 0x18, 0xFF} // ámbar (promedio; ≥4.5:1 sobre panel)
	repHairline  = color.RGBA{0xD9, 0xD2, 0xBD, 0xFF} // filete de regla
	repRule      = color.RGBA{0x8A, 0x82, 0x71, 0xFF} // regla media (bordes de control)
)

// reportTheme ajusta el tema claro de Fyne a la paleta del informe.
type reportTheme struct {
	fyne.Theme
}

func newReportTheme() fyne.Theme {
	return &reportTheme{Theme: theme.LightTheme()}
}

func (t *reportTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameBackground:
		return repPaper
	case theme.ColorNameForeground:
		return repInk
	case theme.ColorNamePrimary:
		return repHydro
	case theme.ColorNameHover:
		return repWash
	case theme.ColorNameButton:
		return repPanel
	case theme.ColorNameDisabledButton:
		return repPanel
	case theme.ColorNameDisabled:
		return repInkMuted
	case theme.ColorNamePlaceHolder:
		return repInkMuted
	case theme.ColorNameError:
		return repVermilion
	case theme.ColorNameSuccess:
		return repMoss
	case theme.ColorNameWarning:
		return repAmber
	case theme.ColorNameInputBackground:
		return repPaper
	case theme.ColorNameInputBorder:
		return repRule
	case theme.ColorNameSeparator:
		return repHairline
	case theme.ColorNameMenuBackground:
		return repPaper
	case theme.ColorNameOverlayBackground:
		return repPaper
	case theme.ColorNameShadow:
		return color.RGBA{0x22, 0x31, 0x3A, 0x28}
	case theme.ColorNameSelection:
		return repHydroSoft
	case theme.ColorNameFocus:
		return repHydro
	case theme.ColorNamePressed:
		return repWash
	case theme.ColorNameScrollBar:
		return repHairline
	case theme.ColorNameScrollBarBackground:
		return repPaper
	default:
		return t.Theme.Color(n, v)
	}
}
