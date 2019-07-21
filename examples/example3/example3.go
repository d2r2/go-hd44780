package main

import (
	"log"
	"strings"
	"time"

	device "github.com/adrianh-za/go-hd44780"
	"github.com/d2r2/go-i2c"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	i2c, err := i2c.NewI2C(0x27, 1)
	checkError(err)
	defer i2c.Close()
	lcd, err := device.NewLcd(i2c, device.LCD_20x4)
	//lcd.SetStrobeDelays(400, 50)
	checkError(err)
	err = lcd.BacklightOn()
	checkError(err)
	var msg = []string{
		"--=! Let's rock !=--",
		"Welcome to RPi dude!",
		"<! I know kung fu !>",
		"R2D2, where are you?",
	}
	lines := []device.ShowOptions{device.SHOW_LINE_1, device.SHOW_LINE_2,
		device.SHOW_LINE_3, device.SHOW_LINE_4}
	i := 0
	for {
		var j byte
		for j = 0; j < 4; j++ {
			k := (i + int(j)) % 4
			err = lcd.ShowMessage(msg[k], lines[j])
			checkError(err)
		}
		time.Sleep(2 * time.Second)
		for j = 0; j < 4; j++ {
			err = lcd.ShowMessage(strings.Repeat(" ", 20), lines[j])
			checkError(err)
		}
		time.Sleep(2 * time.Second)
		//err = lcd.BacklightOff()
		//checkError(err)
		//time.Sleep(2 * time.Second)
		//err = lcd.BacklightOn()
		//checkError(err)
		
		i++
	}
}
