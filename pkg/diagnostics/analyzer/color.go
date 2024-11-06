package analyzer

import "os"

const (
	resetC   = "\033[0m"
	blackC   = "\033[30m"
	redC     = "\033[31m"
	greenC   = "\033[32m"
	yellowC  = "\033[33m"
	blueC    = "\033[34m"
	purpleC  = "\033[35m"
	cyanC    = "\033[36m"
	greyC    = "\033[37m"
	whiteC   = "\033[97m"
	magentaC = "\033[95m"

	underlineC     = "\033[4m"
	resetUnderline = "\033[24m"

	boldC     = "\033[1m"
	resetBold = "\033[22m"
)

var colorDisabled = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"

func noColor() bool {
	return colorDisabled
}

func wrap(m, c, reset string) string {
	if noColor() {
		return m
	}
	return c + m + reset
}

func color(m, c string) string {
	return wrap(m, c, resetC)
}

func blue(m string) string {
	return color(m, blueC)
}

func cyan(m string) string {
	return color(m, cyanC)
}

func red(m string) string {
	return color(m, redC)
}

func green(m string) string {
	return color(m, greenC)
}

func yellow(m string) string {
	return color(m, yellowC)
}

func black(m string) string {
	return color(m, blackC)
}

func grey(m string) string {
	return color(m, greyC)
}

func magenta(m string) string {
	return color(m, magentaC)
}

func underline(m string) string {
	return wrap(m, underlineC, resetUnderline)
}

func bold(m string) string {
	return wrap(m, boldC, resetBold)
}
