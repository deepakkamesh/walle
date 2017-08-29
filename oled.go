package walle

import (
	"errors"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/golang/glog"
	"gobot.io/x/gobot/drivers/i2c"
	"gobot.io/x/gobot/platforms/raspi"
)

type imageData struct {
	Xmax int
	Ymax int
	data [][]bool
}

type Animation struct {
	images []image.Image
}

type OLED struct {
	quitLoop chan struct{}
	curr     uint
	images   []imageData
	tick     *time.Ticker
	updateCh chan *Animation
	oled     *i2c.SSD1306Driver
}

// New returns an initialized OLED.
func NewOLED() *OLED {
	return &OLED{
		quitLoop: make(chan struct{}),
		tick:     time.NewTicker(100 * time.Millisecond),
		curr:     0,
		updateCh: make(chan *Animation),
	}
}

// Init initializes the OLED display.
func (s *OLED) Init(r *raspi.Adaptor, bus int, i2cAddress int) error {

	// Initialize I2C display.
	oled := i2c.NewSSD1306Driver(r, i2c.WithBus(bus), i2c.WithAddress(i2cAddress))
	if err := oled.Start(); err != nil {
		return err
	}
	s.oled = oled
	s.oled.Clear()
	s.oled.SetContrast(10)
	return nil
}

// Animate sends image data to the main processing loop. This is done
// in the main loop to avoid race conditions; updating image data while
// its being displayed by draw func.
func (s *OLED) Animate(imgs []image.Image, d time.Duration) {

	s.updateCh <- &Animation{
		images: imgs,
	}
}

// processImage processes the image data and loads it. It currently only
// processes the A of rgbA of a monochrome image. 'A' indicates the opacity
// of the pixel.
func (s *OLED) processImages(imgs []image.Image, d time.Duration) {

	s.images = nil

	for _, img := range imgs {

		// Allocate Array.
		data := make([][]bool, img.Bounds().Max.Y)
		for j := range data {
			data[j] = make([]bool, img.Bounds().Max.X)
		}

		// Mark X,Y coordinates which are opaque
		for y := 0; y < img.Bounds().Max.Y; y++ {
			for x := 0; x < img.Bounds().Max.X; x++ {
				_, _, _, a := img.At(x, y).RGBA()
				if a > 0 {
					data[y][x] = true
					continue
				}
				data[y][x] = false
			}
		}

		s.images = append(s.images, imageData{
			Xmax: img.Bounds().Max.X,
			Ymax: img.Bounds().Max.Y,
			data: data,
		})
	}
}

// draw displays image on OLED display.
func (s *OLED) draw() {
	if len(s.images) == 0 {
		return
	}
	s.oled.Clear()
	w := s.oled.Buffer.Width
	h := s.oled.Buffer.Height

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if s.images[s.curr].Ymax > y && s.images[s.curr].Xmax > x && s.images[s.curr].data[y][x] {
				s.oled.Set(x, y, 1)
			}
		}
	}
	if err := s.oled.Display(); err != nil {
		glog.Errorf("Failed to display:%v", err)
	}
}

func (s *OLED) Run() error {
	if s == nil {
		errors.New("OLED not initialized")
	}

	i := 0

	go func() {
		for {
			select {
			case upd := <-s.updateCh:
				i = 0
				s.processImages(upd.images, 200)

			case <-s.tick.C:
				if i == len(s.images) {
					i = 0
				}
				s.curr = uint(i)
				s.draw()
				i++

			case <-s.quitLoop:
				s.oled.Clear()
				return
			}
		}
	}()

	return nil
}

func (s *OLED) Quit() {
	// Called in a goroutine because of potential deadlock with
	// poller loop.
	go func() {
		s.quitLoop <- struct{}{}
	}()
}

// LoadImages loads images as a image struct and returns a list.
func LoadImages(images ...string) ([]image.Image, error) {
	imageList := []image.Image{}

	for _, imgFile := range images {
		f, err := os.Open(imgFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		img, err := png.Decode(f)
		if err != nil {
			return nil, err
		}
		imageList = append(imageList, img)
	}
	return imageList, nil
}
