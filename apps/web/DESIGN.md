---
version: alpha
name: Monochrome-Linear-Inspired-Design
description: "A strict monochrome, high-precision design system for ClipIN built around pure black (#000000) canvas, dark neutral surface ladder (#0d0d0d, #121212, #1a1a1a), crisp hairline borders (#262626, #333333), and high-contrast white (#ffffff) primary accents. Engineered for restraint, clarity, and mathematical alignment."

colors:
  primary: "#ffffff"
  on-primary: "#000000"
  primary-hover: "#e5e5e5"
  primary-focus: "#ffffff"
  ink: "#ffffff"
  ink-muted: "#d4d4d4"
  ink-subtle: "#a3a3a3"
  ink-tertiary: "#737373"
  canvas: "#000000"
  surface-1: "#0a0a0a"
  surface-2: "#121212"
  surface-3: "#1a1a1a"
  surface-4: "#262626"
  hairline: "#262626"
  hairline-strong: "#333333"
  hairline-tertiary: "#404040"
  inverse-canvas: "#ffffff"
  inverse-surface-1: "#f5f5f5"
  inverse-surface-2: "#e5e5e5"
  inverse-ink: "#000000"
  semantic-success: "#10b981"
  semantic-overlay: "#000000"

typography:
  display-xl:
    fontFamily: Inter, SF Pro Display, system-ui
    fontSize: 72px
    fontWeight: 600
    lineHeight: 1.1
    letterSpacing: -2.5px
  display-lg:
    fontFamily: Inter, SF Pro Display, system-ui
    fontSize: 48px
    fontWeight: 600
    lineHeight: 1.15
    letterSpacing: -1.5px
  display-md:
    fontFamily: Inter, SF Pro Display, system-ui
    fontSize: 36px
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: -1.0px
  headline:
    fontFamily: Inter, SF Pro Display, system-ui
    fontSize: 24px
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: -0.5px
  body:
    fontFamily: Inter, system-ui
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: -0.05px
  mono:
    fontFamily: JetBrains Mono, monospace
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.5

rounded:
  xs: 4px
  sm: 6px
  md: 8px
  lg: 12px
  pill: 9999px

components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.on-primary}"
    padding: 8px 16px
    height: 36px
    rounded: "{rounded.md}"
  button-secondary:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.ink}"
    border: "1px solid {colors.hairline}"
    padding: 8px 16px
    height: 36px
    rounded: "{rounded.md}"
---

## Overview

ClipIN uses a **strict monochrome, Linear-inspired design system**.
The palette removes chromatic colors (purples, blues) in favor of high-contrast pure black (`#000000`), neutral dark panels (`#0a0a0a`, `#121212`), crisp hairline borders (`#262626`), and solid white (`#ffffff`) as the primary interactive accent.
