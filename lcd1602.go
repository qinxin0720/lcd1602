package lcd1602

import (
	"fmt"
	"log"
	"time"

	"github.com/qinxin0720/go-rpigpio"
)

type AdafruitCharlcd struct {
	//commands
	LCD_CLEARDISPLAY   byte
	LCD_RETURNHOME     byte
	LCD_ENTRYMODESET   byte
	LCD_DISPLAYCONTROL byte
	LCD_CURSORSHIFT    byte
	LCD_FUNCTIONSET    byte
	LCD_SETCGRAMADDR   byte
	LCD_SETDDRAMADDR   byte

	//flags for display entry mode
	LCD_ENTRYRIGHT          byte
	LCD_ENTRYLEFT           byte
	LCD_ENTRYSHIFTINCREMENT byte
	LCD_ENTRYSHIFTDECREMENT byte

	//flags for display on/off control
	LCD_DISPLAYON  byte
	LCD_DISPLAYOFF byte
	LCD_CURSORON   byte
	LCD_CURSOROFF  byte
	LCD_BLINKON    byte
	LCD_BLINKOFF   byte

	//flags for display/cursor shift
	LCD_DISPLAYMOVE byte
	LCD_CURSORMOVE  byte
	LCD_MOVERIGHT   byte
	LCD_MOVELEFT    byte

	//flags for function set
	LCD_8BITMODE byte
	LCD_4BITMODE byte
	LCD_2LINE    byte
	LCD_1LINE    byte
	LCD_5x10DOTS byte
	LCD_5x8DOTS  byte

	//gpio
	pin_rs  int
	pin_e   int
	pins_db [4]int

	//*rpi.Pin
	rs_Pin  *rpi.Pin
	e_Pin   *rpi.Pin
	db_Pins [4]*rpi.Pin

	displaycontrol  byte
	displayfunction byte
	displaymode     byte

	numlines int
	currline int

	row_offsets [4]byte
}

func NewAdafruitCharlcd(pin_rs, pin_e int, pins_db [4]int) *AdafruitCharlcd {
	return &AdafruitCharlcd{
		//commands
		LCD_CLEARDISPLAY:   0x01,
		LCD_RETURNHOME:     0x02,
		LCD_ENTRYMODESET:   0x04,
		LCD_DISPLAYCONTROL: 0x08,
		LCD_CURSORSHIFT:    0x10,
		LCD_FUNCTIONSET:    0x20,
		LCD_SETCGRAMADDR:   0x40,
		LCD_SETDDRAMADDR:   0x80,

		//flags for display entry mode
		LCD_ENTRYRIGHT:          0x00,
		LCD_ENTRYLEFT:           0x02,
		LCD_ENTRYSHIFTINCREMENT: 0x01,
		LCD_ENTRYSHIFTDECREMENT: 0x00,

		//flags for display on/off control
		LCD_DISPLAYON:  0x04,
		LCD_DISPLAYOFF: 0x00,
		LCD_CURSORON:   0x02,
		LCD_CURSOROFF:  0x00,
		LCD_BLINKON:    0x01,
		LCD_BLINKOFF:   0x00,

		//flags for display/cursor shift
		LCD_DISPLAYMOVE: 0x08,
		LCD_CURSORMOVE:  0x00,

		//flags for display/cursor shift
		LCD_MOVERIGHT: 0x04,
		LCD_MOVELEFT:  0x00,

		//flags for function set
		LCD_8BITMODE: 0x10,
		LCD_4BITMODE: 0x00,
		LCD_2LINE:    0x08,
		LCD_1LINE:    0x00,
		LCD_5x10DOTS: 0x04,
		LCD_5x8DOTS:  0x00,

		//gpio
		pin_rs:  pin_rs,
		pin_e:   pin_e,
		pins_db: pins_db,

		//*rpi.Pin
		rs_Pin:  nil,
		e_Pin:   nil,
		db_Pins: [4]*rpi.Pin{nil},

		displaycontrol:  0,
		displayfunction: 0,
		displaymode:     0,
		numlines:        0,
		currline:        0,
		row_offsets:     [4]byte{0, 0, 0, 0},
	}
}

func (a *AdafruitCharlcd) Init() {
	var err error
	a.e_Pin, err = rpi.OpenPin(a.pin_e, rpi.OUT)
	if err != nil {
		log.Fatalf("open GPIO%d error: %s", a.pin_e, err)
	}
	a.rs_Pin, err = rpi.OpenPin(a.pin_rs, rpi.OUT)
	if err != nil {
		log.Fatalf("open GPIO%d error: %s", a.pin_rs, err)
	}
	for i, p := range a.pins_db {
		a.db_Pins[i], err = rpi.OpenPin(p, rpi.OUT)
		if err != nil {
			log.Fatalf("open GPIO%d error: %s", a.pins_db[i], err)
		}
	}

	a.write4bits(0x33, rpi.LOW) //initialization
	a.write4bits(0x32, rpi.LOW) //initialization
	a.write4bits(0x28, rpi.LOW) //2 line 5x7 matrix
	a.write4bits(0x0C, rpi.LOW) //turn cursor off 0x0E to enable cursor
	a.write4bits(0x06, rpi.LOW) //shift cursor right

	a.displaycontrol = a.LCD_DISPLAYON | a.LCD_CURSOROFF | a.LCD_BLINKOFF

	a.displayfunction = a.LCD_4BITMODE | a.LCD_1LINE | a.LCD_5x8DOTS
	a.displayfunction |= a.LCD_2LINE

	//Initialize to default text direction (for romance languages)
	a.displaymode = a.LCD_ENTRYLEFT | a.LCD_ENTRYSHIFTDECREMENT
	a.write4bits(a.LCD_ENTRYMODESET|a.displaymode, rpi.LOW) //set the entry mode

	a.Clear()
}

func (a *AdafruitCharlcd) Begin(cols, lines int) {
	if lines > 1 {
		a.numlines = lines
		a.displayfunction |= a.LCD_2LINE
		a.currline = 0
	}
}

func (a *AdafruitCharlcd) Home() {
	a.write4bits(a.LCD_RETURNHOME, rpi.LOW) //set cursor position to zero
	time.Sleep(3000 * time.Microsecond)     //this command takes a long time!
}

func (a *AdafruitCharlcd) Clear() {
	a.write4bits(a.LCD_CLEARDISPLAY, rpi.LOW) //command to clear display
	time.Sleep(3000 * time.Microsecond)       //3000 microsecond sleep, clearing the display takes a long time
}

func (a *AdafruitCharlcd) SetCursor(col, row int) {
	a.row_offsets = [4]byte{0x00, 0x40, 0x14, 0x54}

	if row > a.numlines {
		row = a.numlines + 1 //we count rows starting w/0
	}

	a.write4bits(a.LCD_SETDDRAMADDR|byte(col+int(a.row_offsets[row])), rpi.LOW)
}

func (a *AdafruitCharlcd) NoDisplay() {
	//Turn the display off (quickly)
	a.displaycontrol &= ^a.LCD_DISPLAYON
	a.write4bits(a.LCD_DISPLAYCONTROL|a.displaycontrol, rpi.LOW)
}

func (a *AdafruitCharlcd) Display() {
	//Turn the display on (quickly)
	a.displaycontrol |= a.LCD_DISPLAYON
	a.write4bits(a.LCD_DISPLAYCONTROL|a.displaycontrol, rpi.LOW)
}

func (a *AdafruitCharlcd) NoCursor() {
	//Turns the underline cursor on/off
	a.displaycontrol &= ^a.LCD_CURSORON
	a.write4bits(a.LCD_DISPLAYCONTROL|a.displaycontrol, rpi.LOW)
}

func (a *AdafruitCharlcd) Cursor() {
	//Cursor On
	a.displaycontrol |= a.LCD_CURSORON
	a.write4bits(a.LCD_DISPLAYCONTROL|a.displaycontrol, rpi.LOW)
}

func (a *AdafruitCharlcd) NoBlink() {
	//Turn on and off the blinking cursor
	a.displaycontrol &= ^a.LCD_BLINKON
	a.write4bits(a.LCD_DISPLAYCONTROL|a.displaycontrol, rpi.LOW)
}

func (a *AdafruitCharlcd) DisplayLeft() {
	//These commands scroll the display without changing the RAM
	a.write4bits(a.LCD_CURSORSHIFT|a.LCD_DISPLAYMOVE|a.LCD_MOVELEFT, rpi.LOW)
}

func (a *AdafruitCharlcd) ScrollDisplayRight() {
	//These commands scroll the display without changing the RAM
	a.write4bits(a.LCD_CURSORSHIFT|a.LCD_DISPLAYMOVE|a.LCD_MOVERIGHT, rpi.LOW)
}

func (a *AdafruitCharlcd) LeftToRight() {
	//This is for text that flows Left to Right
	a.displaymode |= a.LCD_ENTRYLEFT
	a.write4bits(a.LCD_ENTRYMODESET|a.displaymode, rpi.LOW)
}

func (a *AdafruitCharlcd) RightToLeft() {
	//This is for text that flows Right to Left
	a.displaymode &= ^a.LCD_ENTRYLEFT
	a.write4bits(a.LCD_ENTRYMODESET|a.displaymode, rpi.LOW)
}

func (a *AdafruitCharlcd) Autoscroll() {
	//This will 'right justify' text from the cursor
	a.displaymode |= a.LCD_ENTRYSHIFTINCREMENT
	a.write4bits(a.LCD_ENTRYMODESET|a.displaymode, rpi.LOW)
}

func (a *AdafruitCharlcd) NoAutoscroll() {
	//This will 'left justify' text from the cursor
	a.displaymode &= ^a.LCD_ENTRYSHIFTINCREMENT
	a.write4bits(a.LCD_ENTRYMODESET|a.displaymode, rpi.LOW)
}

func (a *AdafruitCharlcd) write4bits(bits byte, charMode rpi.Value) {
	//Send command to LCD
	time.Sleep(1000 * time.Microsecond)

	_bits := fmt.Sprintf("%08b", bits)

	a.rs_Pin.Write(charMode)

	for _, pin := range a.db_Pins {
		pin.Write(rpi.LOW)
	}

	for i := 0; i < 4; i++ {
		if _bits[i] == '1' {
			a.db_Pins[3-i].Write(rpi.HIGH)
		}
	}

	a.pulseEnable()

	for _, pin := range a.db_Pins {
		pin.Write(rpi.LOW)
	}

	for i := 4; i < 8; i++ {
		if _bits[i] == '1' {
			a.db_Pins[7-i].Write(rpi.HIGH)
		}
	}

	a.pulseEnable()
}

func (a *AdafruitCharlcd) pulseEnable() {
	a.e_Pin.Write(rpi.LOW)
	time.Sleep(time.Microsecond) //1 microsecond pause - enable pulse must be > 450ns
	a.e_Pin.Write(rpi.HIGH)
	time.Sleep(time.Microsecond) //1 microsecond pause - enable pulse must be > 450ns
	a.e_Pin.Write(rpi.LOW)
	time.Sleep(time.Microsecond) //commands need > 37us to settle
}

func (a *AdafruitCharlcd) Message(text string) {
	//Send string to LCD. Newline wraps to second line
	for _, char := range []byte(text) {
		if char == '\n' {
			a.write4bits(0xC0, rpi.LOW) //next line
		} else {
			a.write4bits(byte(char), rpi.HIGH)
		}
	}
}
