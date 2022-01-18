package e2e

import (
	"fmt"
	"strings"
)

type copyCommand string

func newCopyCommand() copyCommand {
	return "aws s3 cp"
}

func (c copyCommand) String() string {
	return string(c)
}

func (c copyCommand) addOption(opt string) copyCommand {
	return copyCommand(fmt.Sprintf("%s %s", string(c), opt))
}

func (c copyCommand) from(p ...string) copyCommand {
	return c.addOption(strings.Join(p, "/"))
}

func (c copyCommand) to(p ...string) copyCommand {
	return c.addOption(strings.Join(p, "/"))
}

func (c copyCommand) recursive() copyCommand {
	return c.addOption("--recursive")
}

func (c copyCommand) exclude(v string) copyCommand {
	return c.addOption(fmt.Sprintf("--exclude \"%s\"", v))
}

func (c copyCommand) include(v string) copyCommand {
	return c.addOption(fmt.Sprintf("--include \"%s\"", v))
}
