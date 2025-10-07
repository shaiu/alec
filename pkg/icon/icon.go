package icon

// Icons holds all the icon definitions for the application
type Icons struct {
	// Navigation
	ArrowUp    string
	ArrowDown  string
	ArrowLeft  string
	ArrowRight string
	ChevronUp  string
	Parent     string

	// Actions
	Search  string
	Refresh string
	Execute string
	Quit    string

	// Status
	Success string
	Warning string
	Error   string
	Info    string

	// UI Elements
	Lightning string
	Bullet    string
	Separator string
}

var (
	// NerdFontIcons are the icons when Nerd Fonts are available
	NerdFontIcons = Icons{
		// Navigation
		ArrowUp:    "\ueaa1", // nf-cod-arrow_up
		ArrowDown:  "\uea9a", // nf-cod-arrow_down
		ArrowLeft:  "\uea9b", // nf-cod-arrow_left
		ArrowRight: "\uea9c", // nf-cod-arrow_right
		ChevronUp:  "\ueab7", // nf-cod-chevron_up
		Parent:     "\ueab7", // nf-cod-chevron_up (for parent directory)

		// Actions
		Search:  "\uea6d", // nf-cod-search
		Refresh: "\ueb37", // nf-cod-refresh
		Execute: "\ueb2c", // nf-cod-play
		Quit:    "\uea76", // nf-cod-close

		// Status
		Success: "\ueab2", // nf-cod-check
		Warning: "\uea6c", // nf-cod-warning
		Error:   "\uf00d", // nf-fa-times
		Info:    "\uea74", // nf-cod-info

		// UI Elements
		Lightning: "\uf0e7", // nf-fa-bolt
		Bullet:    "\u2022", // bullet point (standard Unicode)
		Separator: "\u2022", // bullet point for separators
	}

	// ASCIIIcons are fallback icons for terminals without Nerd Fonts
	ASCIIIcons = Icons{
		// Navigation
		ArrowUp:    "↑",
		ArrowDown:  "↓",
		ArrowLeft:  "←",
		ArrowRight: "→",
		ChevronUp:  "^",
		Parent:     "↑",

		// Actions
		Search:  "/",
		Refresh: "r",
		Execute: ">",
		Quit:    "q",

		// Status
		Success: "✓",
		Warning: "!",
		Error:   "✗",
		Info:    "i",

		// UI Elements
		Lightning: "⚡",
		Bullet:    "•",
		Separator: "•",
	}

	// Current holds the active icon set
	Current = NerdFontIcons
)

// UseNerdFont sets the icon set to Nerd Font icons
func UseNerdFont() {
	Current = NerdFontIcons
}

// UseASCII sets the icon set to ASCII fallback icons
func UseASCII() {
	Current = ASCIIIcons
}
