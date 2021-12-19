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

type InitCommand struct {
	fs  *flag.FlagSet
	spi string
	dc  string
}

func NewInitCommand() *InitCommand {
	c := &InitCommand{
		fs: flag.NewFlagSet("init", flag.ExitOnError),
	}

	c.fs.StringVar(&c.spi, "spi", "", "SPI device (SPI0.1)")
	c.fs.StringVar(&c.dc, "dc", "", "dc pin (9)")

	return c
}

func (c *InitCommand) Name() string {
	return c.fs.Name()
}

func (c *InitCommand) Init(args []string) error {
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

	return nil
}

func (c *InitCommand) Execute() error {
	fmt.Printf("Initializing")

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

	dev, err := st7735.New(conn, dcPin, nil, nil, &st7735.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}

	return dev.Init()
}
