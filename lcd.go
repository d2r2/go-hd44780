package hd44780

import (
	"fmt"
	"strings"
	"time"

	"github.com/d2r2/go-i2c"
)

const (
	// Commands
	CMD_Clear_Display        = 0x01
	CMD_Return_Home          = 0x02
	CMD_Entry_Mode           = 0x04
	CMD_Display_Control      = 0x08
	CMD_Cursor_Shift		 = 0x10
	CMD_Function_Set         = 0x20
	CMD_CGRAM_Set            = 0x40
	CMD_DDRAM_Set            = 0x80

	// Flags for display entry mode (CMD_Entry_Mode)
	OPT_EntryLeft				= 0x02
	OPT_EntryRight				= 0x00
	OPT_Increment 				= 0x01 
	OPT_Decrement 				= 0x00

	// Flags for display control (CMD_Display_Control)
	OPT_Enable_Display 	= 0x04
	OPT_Disable_Display = 0x00
	OPT_Enable_Cursor  	= 0x02
	OPT_Disable_Cursor  = 0x00
	OPT_Enable_Blink   	= 0x01
	OPT_Disable_Blink   = 0x00
	
	// Flags for display/cursor move ()
	OPT_Display_Move  	= 0x08
	OPT_Cursor_Move		= 0x00
	OPT_Move_Right    	= 0x04
	OPT_Move_Left    	= 0x00
	
	// Flags for function set (CMD_Function_Set)
	OPT_8Bit_Mode      = 0x10
	OPT_4Bit_Mode      = 0x00
	OPT_2_Lines        = 0x08
	OPT_1_Lines        = 0x00
	OPT_5x10_Dots      = 0x04
	OPT_5x8_Dots       = 0x00

	PIN_BACKLIGHT byte = 0x08
	PIN_EN        byte = 0x04 // Enable bit
	PIN_RW        byte = 0x02 // Read/Write bit
	PIN_RS        byte = 0x01 // Register select bit
)

type LcdType int

const (
	LCD_UNKNOWN LcdType = iota
	LCD_16x2
	LCD_20x4
)

type ShowOptions int

const (
	SHOW_NO_OPTIONS ShowOptions = 0
	SHOW_LINE_1                 = 1 << iota
	SHOW_LINE_2
	SHOW_LINE_3
	SHOW_LINE_4
	SHOW_ELIPSE_IF_NOT_FIT
	SHOW_BLANK_PADDING
)

type Lcd struct {
	i2c       			*i2c.I2C
	backlight 			bool
	lcdType   			LcdType
	writeStrobeDelay	uint16
	resetStrobeDelay	uint16
	active				bool
	displayFunction 	byte
	displayControl  	byte
	displayMode     	byte
}

func NewLcd(i2c *i2c.I2C, lcdType LcdType) (*Lcd, error) {
	this := &Lcd{i2c: i2c,
		backlight: false,
		lcdType: lcdType,
		writeStrobeDelay: 200,
		resetStrobeDelay: 30,
		active: true,
		displayFunction: 0x00,
		displayControl: 0x00,
		displayMode: 0x00,
		}
	
	// Wait is required during initialization steps.  Various info below about delays.
	// https://www.sparkfun.com/datasheets/LCD/HD44780.pdf (page 45)
	// https://github.com/mrmorphic/hwio/blob/master/devices/hd44780/hd44780_i2c.go
	// https://github.com/duinoWitchery/hd44780/blob/master/hd44780.cpp (read the comments)

	// Initial delay as per datasheet (need at least 40ms after power rises above 2.7V before sending commands.)
	time.Sleep(100 * time.Millisecond)	// Wait 100ms vs 40ms

	// Step 1 -> Base initialization sent with safe minimum delay afterwards
	var err = this.writeByte(0x03, 0)	
	if err != nil {
		return nil, err
	}
	time.Sleep(5 * time.Millisecond) 	// Wait 5ms vs 4.1ms

	// Step 2 -> Base initialization sent with safe minimum delay afterwards
	err = this.writeByte(0x03, 0)
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Millisecond) 	// Wait 1ms vs 100us

	// Step 3 -> Base initialization sent with safe minimum delay afterwards
	err = this.writeByte(0x03, 0)
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Millisecond) 	// Wait 1ms vs 100us

	// Step 4 -> 4-bit transfer mode sent with safe minimum delay afterwards
	err = this.writeByte(0x02, 0)
	if err != nil {
		return nil, err
	}
	time.Sleep(1 * time.Millisecond) 	// Wait 1ms vs 100us
	
	// Step 5a -> Execute FUNCTIONSET command
	this.displayFunction = OPT_2_Lines | OPT_5x8_Dots | OPT_4Bit_Mode
	err = this.writeByte(CMD_Function_Set | this.displayFunction, 0)
	time.Sleep(1 * time.Millisecond)	// Wait 1ms	to be safe
	if err != nil {
		return nil, err
	}

	// Step 5b -> Execute DISPLAYCONTROL command
	this.displayControl = OPT_Enable_Display | OPT_Disable_Cursor | OPT_Disable_Blink
	err = this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(1 * time.Millisecond)	// Wait 1ms	to be safe
	if err != nil {
		return nil, err
	}

	// Step 5c -> Execute ENTRYMODE command
	this.displayMode =  OPT_EntryLeft
	err = this.writeByte(CMD_Entry_Mode | this.displayMode, 0)
	time.Sleep(1 * time.Millisecond)	// Wait 1ms	to be safe
	if err != nil {
		return nil, err
   
	}

	// Clear the display
	err = this.Clear()
	if err != nil {
		return nil, err
	}

	// Send cursor to home
	err = this.Home()
	if err != nil {
		return nil, err
	}

	return this, nil
}

type rawData struct {
	Data  byte
	Delay time.Duration
}

func (this *Lcd) writeRawDataSeq(seq []rawData) error {
	for _, item := range seq {
		_, err := this.i2c.WriteBytes([]byte{item.Data})
		if err != nil {
			return err
		}
		time.Sleep(item.Delay)
	}
	return nil
}

func (this *Lcd) writeDataWithStrobe(data byte) error {
	if this.backlight {
		data |= PIN_BACKLIGHT
	}
	seq := []rawData{
		{data, 0}, // send data
		{data | PIN_EN, time.Duration(this.writeStrobeDelay) * time.Microsecond}, 	// set strobe
		{data, time.Duration(this.resetStrobeDelay) * time.Microsecond},          	// reset strobe
	}
	return this.writeRawDataSeq(seq)
}

func (this *Lcd) writeByte(data byte, controlPins byte) error {
	err := this.writeDataWithStrobe(data&0xF0 | controlPins)
	if err != nil {
		return err
	}
	err = this.writeDataWithStrobe((data<<4)&0xF0 | controlPins)
	if err != nil {
		return err
	}
	return nil
}

func (this *Lcd) getLineRange(options ShowOptions) (startLine, endLine int) {
	var lines [4]bool
	lines[0] = options&SHOW_LINE_1 != 0
	lines[1] = options&SHOW_LINE_2 != 0
	lines[2] = options&SHOW_LINE_3 != 0
	lines[3] = options&SHOW_LINE_4 != 0
	startLine = -1
	for i := 0; i < len(lines); i++ {
		if lines[i] {
			startLine = i
			break
		}
	}
	endLine = -1
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] {
			endLine = i
			break
		}
	}
	return startLine, endLine
}

func (this *Lcd) splitText(text string, options ShowOptions) []string {
	var lines []string
	startLine, endLine := this.getLineRange(options)
	w, _ := this.getSize()
	if w != -1 && startLine != -1 && endLine != -1 {
		for i := 0; i <= endLine-startLine; i++ {
			if len(text) == 0 {
				break
			}
			j := w
			if j > len(text) {
				j = len(text)
			}
			lines = append(lines, text[:j])
			text = text[j:]
		}
		if len(text) > 0 {
			if options&SHOW_ELIPSE_IF_NOT_FIT != 0 {
				j := len(lines) - 1
				lines[j] = lines[j][:len(lines[j])-1] + "~"
			}
		} else {
			if options&SHOW_BLANK_PADDING != 0 {
				j := len(lines) - 1
				lines[j] = lines[j] + strings.Repeat(" ", w-len(lines[j]))
				for k := j + 1; k <= endLine-startLine; k++ {
					lines = append(lines, strings.Repeat(" ", w))
				}
			}

		}
	} else if len(text) > 0 {
		lines = append(lines, text)
	}
	return lines
}

func (this *Lcd) ShowMessage(text string, options ShowOptions) error {
	//Not active, so don't try do anything
	if (!this.active) {
		return nil
	}

	lines := this.splitText(text, options)
	log.Debug("Output: %v\n", lines)
	startLine, endLine := this.getLineRange(options)
	i := 0
	for {
		if startLine != -1 && endLine != -1 {
			err := this.SetPosition(i+startLine, 0)
			if err != nil {
				return err
			}
		}
		line := lines[i]
		for _, c := range line {
			err := this.writeByte(byte(c), PIN_RS)
			if err != nil {
				return err
			}
		}
		if i == len(lines)-1 {
			break
		}
		i++
	}
	return nil
}

func (this *Lcd) TestWriteCGRam() error {
	err := this.writeByte(CMD_CGRAM_Set, 0)
	if err != nil {
		return err
	}
	var a byte = 0x55
	for i := 0; i < 80; i++ {
		err := this.writeByte(a, PIN_RS)
		if err != nil {
			return err
		}
		a = a ^ 0xFF
	}
	return nil
}

func (this *Lcd) BacklightOn() error {
	this.backlight = true
	err := this.writeByte(0x00, 0)
	time.Sleep(100 * time.Millisecond)		// Used delay as specified in https://github.com/duinoWitchery/hd44780/blob/master/hd44780.cpp (Line 885)
	if err != nil {
		return err
	}
	return nil
}

func (this *Lcd) BacklightOff() error {
	this.backlight = false
	err := this.writeByte(0x00, 0)
	time.Sleep(250 * time.Millisecond)		// Used delay as specified in https://github.com/duinoWitchery/hd44780/blob/master/hd44780.cpp (Line 887)
	if err != nil {
		return err
	}
	return nil
}

func (this *Lcd) Clear() error {
	err := this.writeByte(CMD_Clear_Display, 0)
	time.Sleep(2 * time.Millisecond)		// Do same delay as Home().
	return err
}

func (this *Lcd) Home() error {
	err := this.writeByte(CMD_Return_Home, 0)
	time.Sleep(2 * time.Millisecond)		// Page 24 of datasheet says 1.52ms to execute.  We will do slightly longer delay.
	return err
}

func (this *Lcd) DisplayOn() error {
	this.displayControl  |= OPT_Enable_Display
	err := this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) DisplayOff() error {
	this.displayControl  = this.displayControl  &^ OPT_Enable_Display
	err := this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) BlinkOn() error {
	this.displayControl |= OPT_Enable_Blink
	err := this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) BlinkOff() error {
	this.displayControl = this.displayControl &^ OPT_Enable_Blink
	err := this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) CursorOn() error {
	this.displayControl |= OPT_Enable_Cursor
	err := this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) CursorOff() error {
	this.displayControl = this.displayControl &^ OPT_Enable_Cursor
	err := this.writeByte(CMD_Display_Control | this.displayControl, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) ScrollDisplayLeft() error {
	err := this.writeByte(CMD_Cursor_Shift | OPT_Display_Move | OPT_Move_Left, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) ScrollDisplayRight() error {
	err := this.writeByte(CMD_Cursor_Shift | OPT_Display_Move | OPT_Move_Right, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) LeftRightDisplay() error {
	this.displayMode |= OPT_EntryLeft
	err := this.writeByte(CMD_Entry_Mode | this.displayMode, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) RightLeftDisplay() error {
	this.displayMode = this.displayMode &^ OPT_EntryLeft
	err := this.writeByte(CMD_Entry_Mode | this.displayMode, 0)
	time.Sleep(2 * time.Millisecond)        // Do same delay as Home().
	return err
}

func (this *Lcd) getSize() (width, height int) {
	switch this.lcdType {
	case LCD_16x2:
		return 16, 2
	case LCD_20x4:
		return 20, 4
	default:
		return -1, -1
	}
}

func (this *Lcd) SetPosition(line, pos int) error {
	//Not active, so don't try do anything
	if (!this.active) {
		return nil
	}
	
	w, h := this.getSize()
	if w != -1 && (pos < 0 || pos > w-1) {
		return fmt.Errorf("Cursor position %d "+
			"must be within the range [0..%d]", pos, w-1)
	}
	if h != -1 && (line < 0 || line > h-1) {
		return fmt.Errorf("Cursor line %d "+
			"must be within the range [0..%d]", line, h-1)
	}
	lineOffset := []byte{0x00, 0x40, 0x14, 0x54}
	var b byte = CMD_DDRAM_Set + lineOffset[line] + byte(pos)
	err := this.writeByte(b, 0)
	return err
}

func (this *Lcd) Write(buf []byte) (int, error) {
	for i, c := range buf {
		err := this.writeByte(c, PIN_RS)
		if err != nil {
			return i, err
		}
	}
	return len(buf), nil
}

func (this *Lcd) Command(cmd byte) error {
	err := this.writeByte(cmd, 0)
	return err
}

// GetStrobeDelays returns the WRITE and RESET strobe delays in microseconds.
func (this *Lcd) GetStrobeDelays() (writeDelay, resetDelay uint16) {
	return this.writeStrobeDelay, this.resetStrobeDelay
}

// SetStrobeDelays sets the WRITE and RESET strobe delays in microseconds.
func (this *Lcd) SetStrobeDelays(writeDelay, resetDelay uint16) {
	this.writeStrobeDelay = writeDelay
	this.resetStrobeDelay = resetDelay
}

// Fill will show the specified character across the entire display
func (this *Lcd) Fill(char rune) (error) {
	//Not active, so don't try do anything
	if (!this.active) {
		return nil
	}
	
	var width, height = this.getSize()

	// Invalid srceen size, do nothing
	if (width * height <= 1) {
		return nil
	}

	// Fill the display line by line
	for lineCount := 0; lineCount < height; lineCount++ {
		// Move cursor to position
		err := this.SetPosition(lineCount, 0)
		if err != nil {
			return err
		}

		// Fill the line
		for colCount := 0; colCount < width; colCount++ {
			err = this.writeByte(byte(char), PIN_RS)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Shutdown will cleanup the LCD display
func (this *Lcd) Shutdown() {
	this.active = false						//Set active to FALSE.  This will "block" characters being written to display (check functions which check this flag)
	time.Sleep(250 * time.Millisecond)		//Sleep to allow for any instructions/commands to complete before we continue
	
	// Shutdown display
	this.BacklightOff()
	this.Clear()
	this.Home()
}