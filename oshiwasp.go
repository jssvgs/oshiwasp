package main

import (
    "github.com/mrmorphic/hwio"
    "fmt"
    "time"
    "os"
    "net/http"
    "regexp"
    "errors"
    "html/template"
)

type Page struct {
    Title string
    Body string
}

const tmplPath = "tmpl/" // path of the template files .html in the local file system
const dataPath = "data/" // path of the data files in the local file system
const dataFileName = "output.csv" //  data file name in the local file system

var templates = template.Must(template.ParseGlob(tmplPath+"*.html"))
var validPath = regexp.MustCompile("^/(index|new|start|pause|resume|stop|download|data)/([a-zA-Z0-9]+)$")

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Name")
    }
    return m[2], nil //the name is the second subexpression
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    theAcq.state = stateNEW
    p := &Page{Title: "Index", Body: "Make an action. State:  "+theAcq.state}
    renderTemplate(w,"index",p)
}

func newHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.state = stateNEW
    // create a new output file
    var e error
    theAcq.outputFile, e = os.Create(theAcq.outputFileName)
    if e != nil {
        panic(e)
    }

    p := &Page{Title: "Index", Body: "New acquisition ready. Select Start to begin it. State: " + theAcq.state}

    renderTemplate(w,"index",p)
}

func startHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.state = stateRUNNING

    //waitTillButtonPushed(buttonA)
    p := &Page{Title: "Start", Body: "State: "+theAcq.state}
    hwio.DigitalWrite(theOshi.statusLed, hwio.HIGH)
    t0 = time.Now()
    fmt.Printf("Beginning.....\n");

    renderTemplate(w,"start",p)

    // launch the trackers

    go readTracker("A", theOshi.trackerA)
    go readTracker("B", theOshi.trackerB)
    go readTracker("C", theOshi.trackerC)
    go readTracker("D", theOshi.trackerD)
}

func stopHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.state = stateSTOPPED
    hwio.DigitalWrite(theOshi.statusLed, hwio.LOW)
    fmt.Printf("Finnishing.....\n");
    // close the GPIO pins
    //hwio.CloseAll()
    theAcq.outputFile.Close() //close the file when finished
    fmt.Print("Finished...")
    p := &Page{Title: "Stop", Body:"State: "+theAcq.state}
    renderTemplate(w,"stop",p)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    p := &Page{Title: "Index", Body: "Download"}
    renderTemplate(w,"download",p)
}

const (
    // setup of the pinout in the raspberry

    statusLedPin = "gpio7" // green      
    actionLedPin = "gpio8"  // yellow     

    buttonAPin = "gpio24"    // start
    buttonBPin = "gpio23"    // stop

    trackerAPin = "gpio22"    
    trackerBPin = "gpio18"    
    trackerCPin = "gpio17"    
    trackerDPin = "gpio4"    
)


const (
    //States for the acquisition
    //                  resume <- PAUSED <- pause
    //                     |                 ^
    //                     |                 |
    //                     +----+        +---+
    //                           \      /
    //0 -- new -> NEW -- start -> RUNNING -- stop -> STOPPED
    //             ^                                   |
    //             |                                   |
    //             +--------------- new ---------------+
    //
    stateNEW = "NEW"
    stateRUNNING = "RUNNING"
    statePAUSED = "PAUSED"
    stateSTOPPED = "STOPPED"
    stateERROR = "ERROR"
)

type Acquisition struct{
    outputFileName string
    outputFile *os.File
    state string
}

func (acq *Acquisition) initiate(){
    acq.state = stateNEW
    acq.outputFileName = dataPath+dataFileName
    var e error
    acq.outputFile, e = os.Create(acq.outputFileName)
    if e != nil {
        panic(e)
    }
}


type Oshiwasp struct {
    statusLed hwio.Pin
    actionLed hwio.Pin
    buttonA hwio.Pin
    buttonB hwio.Pin
    trackerA hwio.Pin
    trackerB hwio.Pin
    trackerC hwio.Pin
    trackerD hwio.Pin
}

func (oshi *Oshiwasp) initiate(){

    var e error
    // Set up 'trakers' as inputs
    oshi.trackerA, e = hwio.GetPinWithMode(trackerAPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    oshi.trackerB, e = hwio.GetPinWithMode(trackerBPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    oshi.trackerC, e = hwio.GetPinWithMode(trackerCPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    oshi.trackerD, e = hwio.GetPinWithMode(trackerDPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
 
    // Set up 'buttons' as inputs
    oshi.buttonA, e = hwio.GetPinWithMode(buttonAPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    oshi.buttonB, e = hwio.GetPinWithMode(buttonBPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }

    // Set up 'leds' as outputs
    oshi.statusLed, e = hwio.GetPinWithMode(statusLedPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }
    oshi.actionLed, e = hwio.GetPinWithMode(actionLedPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }
}

var (

    t0 = time.Now()   // initial time of the loop
    c chan int //channel initialitation
    actionLed hwio.Pin // indicating action in the system

    theAcq=new(Acquisition)
    theOshi=new(Oshiwasp)

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
            dataString := fmt.Sprintf("[%s] %v (%v) -> %d\n",
                       name,timeAction.Sub(t0),timeAction.Sub(timeActionOld),value)
            //fmt.Printf("[%s] %v (%v) -> %d\n",
            //           name,timeAction.Sub(t0),timeAction.Sub(timeActionOld),value)
            fmt.Printf(dataString)
            theAcq.outputFile.WriteString(dataString)
            oldValue = value

            // Write the value to the led indicating somewhat is happened
            if (value == 1) {
                hwio.DigitalWrite(actionLed, hwio.HIGH)
            } else {
                hwio.DigitalWrite(actionLed, hwio.LOW)
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
        // Was the button pressed, value = 1?
        if value == 1 {
            return value
        }
    }
}

func main() {

    // setup 
    mux := http.NewServeMux()
    mux.Handle("/",http.FileServer(http.Dir("data")))
    //mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/index/", indexHandler)
    mux.HandleFunc("/start/", startHandler)
    mux.HandleFunc("/stop/", stopHandler)
    //mux.HandleFunc("/download/",http.FileServer(http.Dir("./data")))

    theAcq.initiate();
    theOshi.initiate();

    // starting the web service...
   // http.Handle("/data", http.FileServer(http.Dir("./data")))
    http.ListenAndServe(":8080", mux)

    // close the GPIO pins
    hwio.CloseAll()
} 
