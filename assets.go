package main

// assets.go — recursos embebidos en el binario (go:embed).
// El binario publicado es autosuficiente: icono y sonidos viajan dentro.
// Los .wav se extraen a ~/.hydrapotion/sounds en el primer arranque porque
// paplay reproduce desde archivo, no desde stdin.

import (
	"embed"
	_ "embed"
)

//go:embed assets/icon.png
var iconPNG []byte

//go:embed assets/add_water.wav assets/popup_appear.wav
var soundFS embed.FS
