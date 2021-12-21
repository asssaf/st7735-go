package st7735

import (
	"errors"
	"image"
	"image/color"
	"time"

	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/spi"
)

const (
	ST7735_SWRESET   = 0x01
	ST7735_SLPIN     = 0x10
	ST7735_SLPOUT    = 0x11
	ST7735_FRMCTR1   = 0xB1
	ST7735_FRMCTR2   = 0xB2
	ST7735_FRMCTR3   = 0xB3
	ST7735_INVCTR    = 0xB4
	ST7735_INVOFF    = 0x20
	ST7735_INVON     = 0x21
	ST7735_PWCTR1    = 0xC0
	ST7735_PWCTR2    = 0xC1
	ST7735_PWCTR4    = 0xC3
	ST7735_PWCTR5    = 0xC4
	ST7735_VMCTR1    = 0xC5
	ST7735_MADCTL    = 0x36
	ST7735_COLMOD    = 0x3A
	ST7735_CASET     = 0x2A
	ST7735_RASET     = 0x2B
	ST7735_RAMWR     = 0x2C
	ST7735_GMCTRP1   = 0xE0
	ST7735_GMCTRN1   = 0xE1
	ST7735_NORON     = 0x13
	ST7735_DISPON    = 0x29
	ST7735_TFTWIDTH  = 80
	ST7735_TFTHEIGHT = 160
	ST7735_COLS      = 132
	ST7735_ROWS      = 162

	ChunkSize = 4096
)

// Dev is a handle to a ST7735
type Dev struct {
	c conn.Conn

	// dc low when sending a command, high when sending data.
	dc gpio.PinOut
	// rst reset pin, active low.
	rst gpio.PinOut
	// backlight pin
	backlight gpio.PinOut

	opts Opts
}

// Opts contains the configuration fot the ST7735 device
type Opts struct {
	Width      byte
	Height     byte
	OffsetLeft byte
	OffsetTop  byte
}

var DefaultOpts = Opts{
	Width:      ST7735_TFTWIDTH,
	Height:     ST7735_TFTHEIGHT,
	OffsetLeft: (ST7735_COLS - ST7735_TFTWIDTH) / 2,  // (132-80)/2 = 26
	OffsetTop:  (ST7735_ROWS - ST7735_TFTHEIGHT) / 2, // (162-160)/2 = 1
}

// New opens a handle to a ST7735
func New(p spi.Port, dc gpio.PinOut, rst gpio.PinOut, backlight gpio.PinOut, o *Opts) (*Dev, error) {
	c, err := p.Connect(4000*physic.KiloHertz, spi.Mode0, 8)

	if err != nil {
		return nil, errors.New("could not connect to device")
	}

	d := &Dev{
		c:         c,
		dc:        dc,
		rst:       rst,
		backlight: backlight,
		opts:      *o,
	}

	return d, nil
}

func (d *Dev) Init() error {
	invert := commandAndData{[]byte{ST7735_INVON}, nil, 0} // don't invert
	if false {
		invert = commandAndData{[]byte{ST7735_INVOFF}, nil, 0} // invert
	}

	width := d.opts.Width
	height := d.opts.Height
	offsetLeft := d.opts.OffsetLeft
	offsetTop := d.opts.OffsetTop

	init := []commandAndData{
		{[]byte{ST7735_SWRESET}, nil, 150 * time.Millisecond},                   // software reset
		{[]byte{ST7735_SLPOUT}, nil, 500 * time.Millisecond},                    // out of sleep mode
		{[]byte{ST7735_FRMCTR1}, []byte{0x01, 0x2C, 0x2D}, 0},                   // frame rate ctrl - normal mode, rate = fosc/(1x2+40) * (LINE+2C+2D)
		{[]byte{ST7735_FRMCTR2}, []byte{0x01, 0x2C, 0x2D}, 0},                   // frame rate crtl - idle mode, rate = fosc(1x2+40) * (LINE+2C+2D)
		{[]byte{ST7735_FRMCTR3}, []byte{0x01, 0x2C, 0x2D, 0x01, 0x2C, 0x2D}, 0}, // frame rate ctrl - partial mode
		{[]byte{ST7735_INVCTR}, []byte{0x07}, 0},                                // display inversion ctrl - no inversion
		{[]byte{ST7735_PWCTR1}, []byte{0xA2, 0x02, 0x84}, 0},                    // power control, -4.6V, auto mode
		{[]byte{ST7735_PWCTR2}, []byte{0x0A, 0x00}, 0},                          // power control, Opamp current small, boost frequency
		{[]byte{ST7735_PWCTR4}, []byte{0x8A, 0x2A}, 0},                          // power control, BCLK/2, Opamp current small & Medium low
		{[]byte{ST7735_PWCTR5}, []byte{0x8A, 0xEE}, 0},                          // power control
		{[]byte{ST7735_VMCTR1}, []byte{0x0E}, 0},                                // power control
		invert,
		{[]byte{ST7735_MADCTL}, []byte{0xC8}, 0},                                          // memory access control (directions), row addr/col addr, bottom to top refresh
		{[]byte{ST7735_COLMOD}, []byte{0x05}, 0},                                          // set color mode, 16-bit color
		{[]byte{ST7735_CASET}, []byte{0x00, offsetLeft, 0x00, width + offsetLeft - 1}, 0}, // Column addr set, XSTART = 0, XEND = ROWS-height
		{[]byte{ST7735_RASET}, []byte{0x00, offsetTop, 0x00, height + offsetTop - 1}, 0},  // Row addr set, XSTART = 0, XEND = COLS-width
		{[]byte{ST7735_GMCTRP1}, []byte{0x02, 0x1c, 0x07, 0x12, // set gamma
			0x37, 0x32, 0x29, 0x2d, 0x29, 0x25, 0x2B, 0x39, 0x00,
			0x01, 0x03, 0x10}, 0},
		{[]byte{ST7735_GMCTRN1}, []byte{0x03, 0x1d, 0x07, 0x06, // set gamma
			0x2E, 0x2C, 0x29, 0x2D, 0x2E, 0x2E, 0x37, 0x3F, 0x00,
			0x00, 0x02, 0x10}, 0},
		{[]byte{ST7735_NORON}, nil, 100 * time.Millisecond},  // normal display on
		{[]byte{ST7735_DISPON}, nil, 100 * time.Millisecond}, // display on
	}

	return d.sendBatch(init)
}

func (d *Dev) Powersave() error {
	return d.sendCommand([]byte{ST7735_SLPIN})
}

func (d *Dev) SetBacklight(value bool) {
	if d.backlight == nil {
		return
	}
	if value {
		d.backlight.Out(gpio.High)
	} else {
		d.backlight.Out(gpio.Low)
	}
}

func (d *Dev) SetWindow(x0, y0, x1, y1 int) error {
	y0 += int(d.opts.OffsetTop)
	y1 += int(d.opts.OffsetTop)

	x0 += int(d.opts.OffsetLeft)
	x1 += int(d.opts.OffsetLeft)

	commands := []commandAndData{
		{[]byte{ST7735_CASET}, []byte{byte(x0 >> 8), byte(x0), byte(x1 >> 8), byte(x1)}, 0}, // column addr set
		{[]byte{ST7735_RASET}, []byte{byte(y0 >> 8), byte(y0), byte(y1 >> 8), byte(y1)}, 0}, // row addr set
		{[]byte{ST7735_RAMWR}, nil, 0}, // write to RAM
	}

	return d.sendBatch(commands)

}

func (d *Dev) DisplayImage(x, y int, img *image.RGBA) error {
	offsetBounds := img.Bounds().Add(image.Point{X: x, Y: y})
	maxBounds := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{int(d.opts.Width), int(d.opts.Height)}}
	offsetBounds = offsetBounds.Intersect(maxBounds)
	bounds := offsetBounds.Sub(image.Point{X: x, Y: y})

	err := d.SetWindow(offsetBounds.Min.X, offsetBounds.Min.Y, offsetBounds.Max.X-1, offsetBounds.Max.Y-1)
	if err != nil {
		return err
	}

	subImg := img.SubImage(bounds)

	buf := []byte{}
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			c := subImg.At(x, y)

			cd := RGBATo565(c.(color.RGBA))
			buf = append(buf, byte(cd>>8), byte(cd))
		}
	}

	return d.sendData(buf)
}

func (d *Dev) Display(data []byte) error {
	err := d.SetWindow(0, 0, int(d.opts.Width-1), int(d.opts.Height-1))
	if err != nil {
		return err
	}

	return d.sendData(data)
}

func (d *Dev) Halt() error {
	d.SetBacklight(false)
	return d.Powersave()
}

func (d *Dev) sendCommand(c []byte) error {
	if err := d.dc.Out(gpio.Low); err != nil {
		return err
	}
	return d.c.Tx(c, nil)
}

func (d *Dev) sendData(c []byte) error {
	if err := d.dc.Out(gpio.High); err != nil {
		return err
	}

	for offset := 0; offset < len(c); offset += ChunkSize {
		end := offset + ChunkSize
		if end > len(c) {
			end = len(c)
		}

		if err := d.c.Tx(c[offset:end], nil); err != nil {
			return err
		}
	}

	return nil
}

func (d *Dev) sendBatch(commands []commandAndData) error {
	for _, cad := range commands {
		if err := d.sendCommand(cad.cmd); err != nil {
			return err
		}

		if len(cad.data) == 0 {
			continue
		}

		if err := d.sendData(cad.data); err != nil {
			return err
		}

		if cad.sleep > 0 {
			time.Sleep(cad.sleep)
		}
	}

	return nil
}

type commandAndData struct {
	cmd   []byte
	data  []byte
	sleep time.Duration
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}
