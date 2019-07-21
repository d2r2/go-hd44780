package main

import (
	"fmt"
	"log"
	"time"

	device "github.com/adrianh-za/go-hd44780"
	i2c "github.com/d2r2/go-i2c"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	i2c, err := i2c.NewI2C(0x27, 1)
	check(err)
	defer i2c.Close()
	lcd, err := device.NewLcd(i2c, device.LCD_16x2)
	check(err)
	lcd.BacklightOn()
	lcd.Clear()
	for {
		lcd.Home()
		t := time.Now()
		lcd.SetPosition(0, 0)
		fmt.Fprint(lcd, t.Format("Monday Jan 2"))
		lcd.SetPosition(1, 0)
		fmt.Fprint(lcd, t.Format("15:04:05 2006"))
		//		lcd.SetPosition(4, 0)
		//		fmt.Fprint(lcd, "i2c, VGA, and Go")
		time.Sleep(333 * time.Millisecond)
	}
}
