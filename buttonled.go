package main

import (
    "github.com/mrmorphic/hwio"
    "fmt"
    "time"
)

const (
    buttonPin = "gpio4"    
    ledPin = "gpio7"       
)

func main() {

    //value readed from button, initially set to 0, because the button will not pressed
    oldValue := 0

    // Set up 'button' as an input
    button, e := hwio.GetPinWithMode(buttonPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }

    // Set up 'led' as an output
    led, e := hwio.GetPinWithMode(ledPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }

    fmt.Printf("Beginning.....\n");
    t0:= time.Now() // time 0

    for {
        // Read the button value
        value, e := hwio.DigitalRead(button)
        t1:= time.Now() // time at this point
        if e != nil {
            panic(e)
        }

        // Did value change?
        if value != oldValue {
            fmt.Printf("[%v] %d\n",t1.Sub(t0),value)
            oldValue = value
             
            // Write the value to the led.
            if (value == 1) {
                hwio.DigitalWrite(led, hwio.HIGH)
            } else {
                hwio.DigitalWrite(led, hwio.LOW)
            }
        }

    }
}
