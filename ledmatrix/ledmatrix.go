package ledmatrix

import (
	"fmt"
	"time"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
)

type LEDMatrix struct {
	rconn []*gpio.DirectPinDriver // Row GPIO.
	cconn []*gpio.DirectPinDriver // Coli GPIO.
	data  [][]byte
}

func New(r uint8, c uint8) *LEDMatrix {
	return &LEDMatrix{
		rconn: make([]*gpio.DirectPinDriver, r),
		cconn: make([]*gpio.DirectPinDriver, c),
	}
}

func (s *LEDMatrix) Start(a gobot.Connection, row []string, col []string) error {

	// Initialize row gpio.
	for i, v := range row {
		d := gpio.NewDirectPinDriver(a, v)
		if e := d.Start(); e != nil {
			return fmt.Errorf("failed to start GPIO:%v", e)
		}
		d.DigitalWrite(1)
		s.rconn[i] = d
	}
	// Initialize col gpio.
	for i, v := range col {
		d := gpio.NewDirectPinDriver(a, v)
		if e := d.Start(); e != nil {
			return fmt.Errorf("failed to start GPIO:%v", e)
		}
		d.DigitalWrite(0)
		s.cconn[i] = d
	}
	go s.loop()
	return nil
}

func (s *LEDMatrix) Test() {
	clen := len(s.cconn)
	rlen := len(s.rconn)
	for {
		for j := 0; j < rlen; j++ {
			s.rconn[j].DigitalWrite(0)

			for i := 0; i < clen; i++ {
				if e := s.cconn[i].DigitalWrite(1); e != nil {
					panic(e)
				}

				time.Sleep(500 * time.Millisecond)
				s.cconn[i].DigitalWrite(0)
			}
			s.rconn[j].DigitalWrite(1)
		}
	}
}

func (s *LEDMatrix) Set(data [][]byte) {
	s.data = data
}

func (s *LEDMatrix) loop() {

	clen := len(s.cconn)
	rlen := len(s.rconn)

	defer func() {
		for i := 0; i < clen; i++ {
			s.cconn[i].DigitalWrite(0)
		}

		for i := 0; i < rlen; i++ {
			s.rconn[i].DigitalWrite(1)
		}
	}()

	for {
		if s.data == nil {
			continue
		}

		for r := 0; r < rlen; r++ {
			cnt := 0

			// Set column.
			for c := 0; c < clen; c++ {
				v := s.data[r][c]
				if v == 1 {
					cnt++
				}
				if e := s.cconn[c].DigitalWrite(v); e != nil {
					panic(e)
					//return fmt.Errorf("Write Error on col %v: %v", c, e)
				}
			}

			// only toggle row if we have showed  something.
			if cnt > 0 {
				if e := s.rconn[r].DigitalWrite(0); e != nil {
					panic(e)
					//return fmt.Errorf("Write Error on row %v: %v", r, e)
				}
				time.Sleep(time.Duration(cnt) * 3 * time.Millisecond)
				if e := s.rconn[r].DigitalWrite(1); e != nil {
					panic(e)
					//return fmt.Errorf("Write Off Error on row %v: %v", r, e)
				}
			}
		}
	}
}

func (s *LEDMatrix) Animate(dlist ...[][]byte) {
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		i := 0

		for {
			<-t.C
			bmp := dlist[i]
			s.Set(bmp)
			i++
			if i == len(dlist) {
				i = 0
			}
		}
	}()
}
