package cmd

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"

	"github.com/asssaf/st7735-go/st7735"
)

type BacklightCommand struct {
	fs         *flag.FlagSet
	spi        string
	dc         string
	backlight  string
	action     bool
	actionName string
}

func newBacklightCommand(actionName string, action bool) *BacklightCommand {
	c := &BacklightCommand{
		fs:     flag.NewFlagSet(actionName, flag.ExitOnError),
		action: action,
	}

	c.fs.StringVar(&c.spi, "spi", "", "SPI device (SPI0.1)")
	c.fs.StringVar(&c.dc, "dc", "", "dc pin (9)")
	c.fs.StringVar(&c.backlight, "backlight", "", "backlight pin (25)")

	return c
}

func NewBacklightOnCommand() *BacklightCommand {
	return newBacklightCommand("backlight-on", true)
}

func NewBacklightOffCommand() *BacklightCommand {
	return newBacklightCommand("backlight-off", false)
}

func (c *BacklightCommand) Name() string {
	return c.fs.Name()
}

func (c *BacklightCommand) Init(args []string) error {
	if err := c.fs.Parse(args); err != nil {
		return err
	}

	flag.Usage = c.fs.Usage

	if len(c.spi) == 0 {
		return errors.New("SPI device must be provided with the -spi flag")
	}

	if len(c.dc) == 0 {
		return errors.New("dc pin must be provided with the -dc flag")
	}

	if len(c.backlight) == 0 {
		return errors.New("backlight pin must be provided with the -backlight flag")
	}

	return nil
}

func (c *BacklightCommand) Execute() error {
	// Make sure periph is initialized.
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	conn, err := spireg.Open(c.spi)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	dcPin := gpioreg.ByName(c.dc)
	if dcPin == nil {
		return errors.New(fmt.Sprintf("dc pin not found: %s", c.dc))
	}

	backlightPin := gpioreg.ByName(c.backlight)
	if backlightPin == nil {
		return errors.New(fmt.Sprintf("backlight pin not found: %s", c.backlight))
	}

	dev, err := st7735.New(conn, dcPin, nil, backlightPin, &st7735.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Turning backlight %s\n", c.actionName)

	dev.SetBacklight(c.action)
	return nil
}
