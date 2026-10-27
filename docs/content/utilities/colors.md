---
title: "Colors & Palettes"
weight: 2
---

# Embed Colors & Palettes

Discord embeds feature a colored sidebar strip defined by an integer RGB value (`0xRRGGBB`). Wumpkit includes a typed `Color` primitive and a comprehensive palette of predefined constants.

## Predefined Color Constants

You can pass color constants directly to `SetColor` on embed builders:

```go
package main

import (
	"github.com/wumpkit/wumpkit"
)

func createSuccessEmbed() wumpkit.Embed {
	return wumpkit.NewEmbedBuilder().
		SetTitle("Task Completed").
		SetColor(wumpkit.ColorMediumSeaGreen).
		Build()
}

func createErrorEmbed() wumpkit.Embed {
	return wumpkit.NewEmbedBuilder().
		SetTitle("Action Failed").
		SetColor(wumpkit.ColorRuby).
		Build()
}
```

Frequently used color constants:

| Constant | Hex Value | Color Name |
| :--- | :--- | :--- |
| `ColorBlurple` | `#5865F2` | Discord Blurple |
| `ColorMediumSeaGreen` | `#2ECC71` | Green / Success |
| `ColorSummerSky` | `#3498DB` | Blue / Info |
| `ColorMoonYellow` | `#F1C40F` | Yellow / Warning |
| `ColorTahitiGold` | `#E67E22` | Orange |
| `ColorCinnabar` | `#E74C3C` | Red / Error |
| `ColorRuby` | `#E91E63` | Pink / Crimson |
| `ColorDeepLilac` | `#9B59B6` | Purple |
| `ColorWhite` | `#FFFFFF` | Pure White |
| `ColorBlack` | `#000000` | Pure Black |

## Parsing Hex Strings

If your bot loads color codes from a database or user input (such as `#1ABC9C`), parse it into a `Color`:

```go
color, err := wumpkit.ParseColor("#1ABC9C")
if err != nil {
	// invalid hex string
	return err
}

// format back to standard hex string
hexString := color.String() // "#1ABC9C"
```
