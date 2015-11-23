package main

import (
    "github.com/mrmorphic/hwio"
    "fmt"
    "time"
)

const (
    ledPin = "gpio7"       
    buttonPin1 = "gpio4"    
    buttonPin2 = "gpio17"    
    buttonPin3 = "gpio22"    
)


var (

    t0 = time.Now()   // initial time of the loop
    c chan int //channel initialitation
)



func readButton(name string, buttonPin hwio.Pin, ledPin hwio.Pin){

    //value readed from button, initially set to 0, because the button will not pressed
    oldValue := 0


    t1 := time.Now()

    // loop
    for {
           // Read the button value
           value, e := hwio.DigitalRead(buttonPin)
           if e != nil {
                panic(e)
           }
        t1= time.Now() // time at this point
        // Did value change?
        if value != oldValue {
            fmt.Printf("[%s] %v (%d)\n",name,t1.Sub(t0),value)
            oldValue = value

            // Write the value to the led.
            if (value == 1) {
                hwio.DigitalWrite(ledPin, hwio.HIGH)
            } else {
                hwio.DigitalWrite(ledPin, hwio.LOW)
            }
        }
    }

    
}


func main() {

    // setup 

    // Set up 'button' as an input
    button1, e := hwio.GetPinWithMode(buttonPin1, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    button2, e := hwio.GetPinWithMode(buttonPin2, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    button3, e := hwio.GetPinWithMode(buttonPin3, hwio.INPUT)
    if e != nil {
        panic(e)
    }

    // Set up 'led' as an output
    led, e := hwio.GetPinWithMode(ledPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }

    //e:= hwio.Led("OK", true)
    fmt.Printf("Beginning.....\n");
    t0 = time.Now()
    go readButton("Uno", button1, led)
    go readButton("Dos", button2, led)
    go readButton("Tres", button3, led)

    // wait
    fmt.Print("Enter to finish: ")
    var input string
    fmt.Scanln(&input)
   
    defer hwio.CloseAll()
} 
