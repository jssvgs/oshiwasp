package main

import (
    "github.com/mrmorphic/hwio"
    "fmt"
    "time"
    "os"
    "net/http"
    "io/ioutil"
    "regexp"
    "errors"
    "html/template"
)

type Page struct {
    Title string
    Body []byte
}


var templates = template.Must(template.ParseFiles("index.html", "start.html", "stop.html", "download.html"))
var validPath = regexp.MustCompile("^/(index|start|stop|download)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename,p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    filename := title + ".html"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
       return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string) {
    err := templates.ExecuteTemplate(w, tmpl+".html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

func getName(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("Invalid Page Name")
    }
    return m[2], nil //the name is the second subexpression
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w,"index")
}

func startHandler(w http.ResponseWriter, r *http.Request) {
    // read the button A change to init the data readdings
    //waitTillButtonPushed(buttonA)
    hwio.DigitalWrite(theOshi.statusLed, hwio.HIGH)
    t0 = time.Now()
    fmt.Printf("Beginning.....\n");

    renderTemplate(w,"start")

    // launch the trackers

    go readTracker("A", theOshi.trackerA)
    go readTracker("B", theOshi.trackerB)
    go readTracker("C", theOshi.trackerC)
    go readTracker("D", theOshi.trackerD)

}

func stopHandler(w http.ResponseWriter, r *http.Request) {

    renderTemplate(w,"stop")

    hwio.DigitalWrite(theOshi.statusLed, hwio.LOW)
    fmt.Printf("Finnishing.....\n");

    //fmt.Print("Enter to finish: ")
    //var input string
    //fmt.Scanln(&input)
   

    // close the GPIO pins
    //hwio.CloseAll()

    theAcq.outputFile.Close() //close the file when finished
    fmt.Print("Finished...")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    renderTemplate(w,"download")
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

type Acquisition struct{
    name string
    t0 time.Time
    outputFile *os.File
}

func (acq *Acquisition) Init(name string){
    acq.name=name
    t := time.Now()
    thisOutputFileName := fmt.Sprintf("%s_%d%02d%02d%02d%02d%02d.csv", acq.name,t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
    var e error
    acq.outputFile, e = os.Create(thisOutputFileName)
    if e != nil {
        panic(e)
    }
    //debug
    //fmt.Println("%v", acq)
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

func (oshi *Oshiwasp) Init(){

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
    mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/index", indexHandler)
    mux.HandleFunc("/start", startHandler)
    mux.HandleFunc("/stop", stopHandler)
    mux.HandleFunc("/download", downloadHandler)

    // open file (create if not exists!)
    if len(os.Args) != 2 { 
       fmt.Printf("Usage: %s fileBaseName\n", os.Args[0])
       os.Exit(1)
    }

    theAcq.Init(os.Args[1]);
    theOshi.Init();

    // starting the web service...
    http.ListenAndServe(":8080", mux)

    // close the GPIO pins
    hwio.CloseAll()
} 
