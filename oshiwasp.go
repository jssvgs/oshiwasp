package main

import (
    "github.com/mrmorphic/hwio"
    "fmt"
    "time"
)

const (
    statusLedPin = "gpio7" // green      
    actionLedPin = "gpio8"  // yellow     

    buttonAPin = "gpio24"    // start
    buttonBPin = "gpio23"    // stop

    trackerAPin = "gpio22"    
    trackerBPin = "gpio18"    
    trackerCPin = "gpio17"    
    trackerDPin = "gpio4"    
)


var (

    t0 = time.Now()   // initial time of the loop
    c chan int //channel initialitation
    ledAction hwio.Pin // indicating action in the system
)



func readTracker(name string, trackerPin hwio.Pin){

    //value readed from tracker, initially set to 0, because the tracker was innactive
    oldValue := 0

    timeAction := time.Now() // time of the action detected
    timeActionOld := time.Now() // time of the action-1 detected

    // loop
    for {
           // Read the tracker value
           value, e := hwio.DigitalRead(trackerPin)
           if e != nil {
                panic(e)
           }
        timeActionOld=timeAction //store the last time
        timeAction= time.Now() // time at this point
        // Did value change?
        if value != oldValue {
            fmt.Printf("[%s] %v (%v) -> %d\n",
                       name,timeAction.Sub(t0),timeAction.Sub(timeActionOld),value)
            oldValue = value

            // Write the value to the led indicating somewhat is happened
            if (value == 1) {
                hwio.DigitalWrite(ledAction, hwio.HIGH)
            } else {
                hwio.DigitalWrite(ledAction, hwio.LOW)
            }
        }
    }
}

func waitTillButtonPushed(buttonPin hwio.Pin) int {

    // loop
    for {
        // Read the tracker value
        value, e := hwio.DigitalRead(buttonPin)
        if e != nil {
             panic(e)
        }
        // Did the button pressed, value = 1?
        if value == 1 {
            return value
        }
    }
}

func main() {

    // setup 

    // Set up 'trakers' as inputs
    trackerA, e := hwio.GetPinWithMode(trackerAPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    trackerB, e := hwio.GetPinWithMode(trackerBPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    trackerC, e := hwio.GetPinWithMode(trackerCPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    trackerD, e := hwio.GetPinWithMode(trackerDPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
 
    // Set up 'buttons' as inputs
    buttonA, e := hwio.GetPinWithMode(buttonAPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    buttonB, e := hwio.GetPinWithMode(buttonBPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }

    // Set up 'leds' as outputs
    statusLed, e := hwio.GetPinWithMode(statusLedPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }
    actionLed, e := hwio.GetPinWithMode(actionLedPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }
    ledAction=actionLed // to be global accessed by the trackers

    fmt.Printf("Push button A to start, B to finish...\n");

    // read the button A change to init the data readdings
    waitTillButtonPushed(buttonA)
    hwio.DigitalWrite(statusLed, hwio.HIGH)
    t0 = time.Now()
    fmt.Printf("Beginning.....\n");

    // launch the trackers

    go readTracker("A", trackerA)
    go readTracker("B", trackerB)
    go readTracker("C", trackerC)
    go readTracker("D", trackerD)

    // wait till button B is pushed
    waitTillButtonPushed(buttonB)
    hwio.DigitalWrite(statusLed, hwio.LOW)
    fmt.Printf("Finnishing.....\n");

    //fmt.Print("Enter to finish: ")
    //var input string
    //fmt.Scanln(&input)
   

   // close the GPIO pins
    defer hwio.CloseAll()

} 
