package main

import (
	//"bytes"
	"fmt"
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
	lcd, err := device.NewLcd(i2c, device.LCD_16x2)
	checkError(err)
	err = lcd.BacklightOn()
	checkError(err)
	err = lcd.ShowMessage("--=! Let's rock !=--", device.SHOW_LINE_1)
	checkError(err)
	err = lcd.ShowMessage("Welcome to RPi dude!", device.SHOW_LINE_2)
	checkError(err)
	// err = lcd.ShowMessage("I'm lazy to be lazy.", device.SHOW_LINE_3)
	// checkError(err)
	// err = lcd.ShowMessage("R2D2, where are you?", device.SHOW_LINE_4)
	// checkError(err)
	/*time.Sleep(5 * time.Second)
	for i := 0; i <= 12; i++ {
		var buf bytes.Buffer
		for j := 0; j <= 19; j++ {
			buf.Write([]byte{byte(i*20 + j)})
		}
		err = lcd.ShowMessage(buf.String(), device.SHOW_LINE_1)
		checkError(err)
		time.Sleep(1 * time.Second)
	}
	time.Sleep(5 * time.Second)
	err = lcd.TestWriteCGRam()
	checkError(err)
	for i := 0; i <= 12; i++ {
		var buf bytes.Buffer
		for j := 0; j <= 19; j++ {
			buf.Write([]byte{byte(i*20 + j)})
		}
		err = lcd.ShowMessage(buf.String(), device.SHOW_LINE_1)
		checkError(err)
		time.Sleep(1 * time.Second)
	}*/
	lcd.Clear()
	for {
		lcd.Home()
		t := time.Now()
		lcd.SetPosition(0, 0)
		fmt.Fprint(lcd, t.Format("Monday Jan 2"))
		lcd.SetPosition(1, 1)
		fmt.Fprint(lcd, t.Format("15:04:05 2006"))
		//		lcd.SetPosition(4, 0)
		//		fmt.Fprint(lcd, "i2c, VGA, and Go")
		time.Sleep(666 * time.Millisecond)
	}

}
