package main

import (
	"log"
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
	checkError(err)
	err = lcd.BacklightOn()
	checkError(err)
	/*	var msg = []string{
		"--=! Let's rock !=--",
		"Welcome to RPi dude!",
		"I'm lazy to be lazy.",
		"R2D2, where are you?",
	}*/
	err = lcd.ShowMessage("Hello world !!! How are you? How are you",
		device.SHOW_LINE_1|device.SHOW_LINE_2|
			device.SHOW_ELIPSE_IF_NOT_FIT|device.SHOW_BLANK_PADDING)
	checkError(err)
	time.Sleep(3 * time.Second)
	err = lcd.ShowMessage("Welcome to RPi!!!",
		device.SHOW_LINE_1|device.SHOW_LINE_2|
			device.SHOW_ELIPSE_IF_NOT_FIT|device.SHOW_BLANK_PADDING)
	checkError(err)
}
