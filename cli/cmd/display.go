package cmd

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"

	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/host/v3"

	"github.com/asssaf/st7735-go/st7735"
)

type DisplayCommand struct {
	fs         *flag.FlagSet
	spi        string
	dc         string
	color      int
	offsetLeft int
	offsetTop  int
	width      int
	height     int

	colorRGBA color.RGBA
}

func NewDisplayCommand() *DisplayCommand {
	c := &DisplayCommand{
		fs: flag.NewFlagSet("display", flag.ExitOnError),
	}

	c.fs.StringVar(&c.spi, "spi", "", "SPI device (SPI0.1)")
	c.fs.StringVar(&c.dc, "dc", "", "dc pin (9)")
	c.fs.IntVar(&c.color, "color", 0, "Color to set in rgb (0x000000-0xffffff)")
	c.fs.IntVar(&c.offsetLeft, "offset-left", 0, "Offset from the left")
	c.fs.IntVar(&c.offsetTop, "offset-top", 0, "Offset from the top")
	c.fs.IntVar(&c.width, "width", 80, "Width")
	c.fs.IntVar(&c.height, "height", 160, "Height")

	return c
}

func (c *DisplayCommand) Name() string {
	return c.fs.Name()
}

func (c *DisplayCommand) Init(args []string) error {
	if err := c.fs.Parse(args); err != nil {
		return err
	}

	if len(c.spi) == 0 {
		return errors.New("SPI device must be provided with the -spi flag")
	}

	if len(c.dc) == 0 {
		return errors.New("dc pin must be provided with the -dc flag")
	}

	if c.color < 0 || c.color > 0xffffff {
		return errors.New(fmt.Sprintf("Color out of range: %s", c.color))
	}

	c.colorRGBA = color.RGBA{uint8(c.color >> 16), uint8((c.color >> 8) & 0xff), uint8(c.color & 0xff), 0}
	return nil
}

func (c *DisplayCommand) Execute() error {
	fmt.Printf("Sending data to display\n")

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

	bounds := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{c.width, c.height}}
	img := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			img.Set(x, y, c.colorRGBA)
		}
	}

	return dev.DisplayImage(c.offsetLeft, c.offsetTop, img)
}
