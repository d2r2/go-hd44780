package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	sensor "github.com/d2r2/go-dht"
	device "github.com/d2r2/go-hd44780"
	"github.com/d2r2/go-i2c"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writeTime(lcd *device.Lcd, t time.Time, blink bool, m *sync.Mutex) error {
	msg := t.Format("02/1/06 15:04 MST")
	if blink {
		msg = strings.Replace(msg, ":", " ", 1)
	}
	m.Lock()
	defer m.Unlock()
	err := lcd.ShowMessage(msg, device.SHOW_LINE_1|device.SHOW_ELIPSE_IF_NOT_FIT)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	i2c, err := i2c.NewI2C(0x27, 2)
	checkError(err)
	defer i2c.Close()
	lcd, err := device.NewLcd(i2c, device.LCD_20x4)
	checkError(err)
	err = lcd.BacklightOn()
	checkError(err)
	m := &sync.Mutex{}
	blink := false
	go func() {
		//writeTime(lcd, time.Now(), blink, m)
		c := time.Tick(1 * time.Second)
		for t := range c {
			writeTime(lcd, t, blink, m)
			blink = !blink
		}
	}()

	for {
		temp, hum, _, err := sensor.ReadDHTxxWithRetry(sensor.DHT22, 4, true, 5)
		if err == nil {
			m.Lock()
			lcd.ShowMessage(fmt.Sprintf("T: %0.0f*C Hum: %0.0f%%", temp, hum),
				device.SHOW_LINE_2|device.SHOW_ELIPSE_IF_NOT_FIT)
			m.Unlock()
		} else {
			m.Lock()
			lcd.ShowMessage("DHTxx read error",
				device.SHOW_LINE_2|device.SHOW_ELIPSE_IF_NOT_FIT)
			m.Unlock()
		}
		time.Sleep(10 * time.Second)
	}
}
