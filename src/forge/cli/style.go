package cli

import "fmt"

// ANSI 24-bit True Color escape codes
const (
	ColorReset      = "\033[0m"
	ColorBold       = "\033[1m"
	ColorWhite      = "\033[38;2;255;255;255m"
	ColorLightGray  = "\033[38;2;180;180;180m"
	ColorMedGray    = "\033[38;2;140;140;140m"
	ColorDarkGray   = "\033[38;2;90;90;90m"
	ColorAccentBlue = "\033[38;2;90;160;255m"
	ColorSuccess    = "\033[38;2;80;220;120m"
	ColorError      = "\033[38;2;255;100;100m"
	ColorWarning    = "\033[38;2;255;200;80m"
)

// Iconografía Unicode minimalista
const (
	IconSuccess = "✓"
	IconError   = "✗"
	IconWarning = "⚠"
	IconInfo    = "ℹ"
	IconAction  = "→"
	IconBullet  = "◆"
	IconPrompt  = "»"
)

const Banner = ColorBold + ColorWhite + `███████╗ ██████╗ ██████╗  ██████╗ ███████╗
██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝
█████╗  ██║   ██║██████╔╝██║  ███╗█████╗  
██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝  
██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗
╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝` + ColorReset + "\n" + ColorLightGray + "AI-Powered Software Engineering Assistant" + ColorReset

func printSeparator() {
	fmt.Println(ColorDarkGray + "─────────────────────────────────────────────────────────────────" + ColorReset)
}

func styleTitle(text string) string {
	return ColorBold + ColorWhite + IconBullet + " " + text + ColorReset
}

func styleSuccess(text string) string {
	return ColorSuccess + IconSuccess + ColorReset + " " + ColorWhite + text + ColorReset
}

func styleError(text string) string {
	return ColorError + IconError + ColorReset + " " + ColorError + text + ColorReset
}

func styleWarning(text string) string {
	return ColorWarning + IconWarning + ColorReset + " " + ColorWarning + text + ColorReset
}

func styleInfo(text string) string {
	return ColorAccentBlue + IconInfo + ColorReset + " " + ColorLightGray + text + ColorReset
}

func styleAction(text string) string {
	return ColorAccentBlue + IconAction + ColorReset + " " + ColorWhite + text + ColorReset
}

func promptPrefix(text string) string {
	return ColorAccentBlue + IconPrompt + ColorReset + " " + ColorWhite + text + ColorReset
}
