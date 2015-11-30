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
    "log"
    "runtime"
)

type Page struct {
    Title string
    Body string
}


type Acquisition struct{
    outputFileName string
    outputFile *os.File
    state string
}

func (acq *Acquisition) getState() string{
    return acq.state
}
func (acq *Acquisition) setState(s string){
    acq.state = s
    log.Printf("State set to %s\n", acq.state)
}

func (acq *Acquisition) setOutputFileName(s string){
    acq.outputFileName = s
    log.Printf("Output Filename set to %s\n", acq.outputFileName)
}

func (acq *Acquisition) createOutputFile(){
    var e error
    acq.outputFile, e = os.Create(acq.outputFileName)
    if e != nil {
        panic(e)
    }
    log.Printf("Cretated output File %s", acq.outputFileName)
}

func (acq *Acquisition) reopenOutputFile(){
    var e error
    acq.outputFile, e = os.OpenFile(acq.outputFileName,os.O_WRONLY|os.O_APPEND, 0666)
    if e != nil {
        panic(e)
    }
    log.Printf("Reopen output File %s", acq.outputFileName)
}

func (acq Acquisition) closeOutputFile(){ //close the output file 
    acq.outputFile.Close()
    log.Printf("Closed output File %s", acq.outputFileName)
}

func (acq *Acquisition) initiate(){
    acq.setState(stateNEW)
    acq.setOutputFileName(dataPath+dataFileName)
    acq.createOutputFile()
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

const (
    // web related

    tmplPath = "tmpl/" // path of the template files .html in the local file system
    dataPath = "data/" // path of the data files in the local file system
    dataFileName = "output.csv" //  data file name in the local file system


    // setup of the pinout in the raspberry

    statusLedPin = "gpio7" // green      
    actionLedPin = "gpio8"  // yellow     

    buttonAPin = "gpio24"    // start
    buttonBPin = "gpio23"    // stop

    trackerAPin = "gpio22"    
    trackerBPin = "gpio18"    
    trackerCPin = "gpio17"    
    trackerDPin = "gpio4"    

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


var (

    t0 = time.Now()   // initial time of the loop
    c chan int //channel initialitation
    actionLed hwio.Pin // indicating action in the system

    templates = template.Must(template.ParseGlob(tmplPath+"*.html"))
    validPath = regexp.MustCompile("^/(index|new|start|pause|resume|stop|download|data)/([a-zA-Z0-9]+)$")
    theAcq=new(Acquisition)
    theOshi=new(Oshiwasp)

)



func (oshi *Oshiwasp) initiate(){

    var e error
    // Set up 'trakers' as inputs
    oshi.trackerA, e = hwio.GetPinWithMode(trackerAPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as trackerA\n",trackerAPin)

    oshi.trackerB, e = hwio.GetPinWithMode(trackerBPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as trackerB\n",trackerBPin)

    oshi.trackerC, e = hwio.GetPinWithMode(trackerCPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as trackerC\n",trackerCPin)

    oshi.trackerD, e = hwio.GetPinWithMode(trackerDPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as trackerD\n",trackerDPin)

 
    // Set up 'buttons' as inputs
    oshi.buttonA, e = hwio.GetPinWithMode(buttonAPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as buttonA\n",buttonAPin)

    oshi.buttonB, e = hwio.GetPinWithMode(buttonBPin, hwio.INPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as buttonB\n",buttonBPin)

    // Set up 'leds' as outputs
    oshi.statusLed, e = hwio.GetPinWithMode(statusLedPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as statusLed\n",statusLedPin)

    oshi.actionLed, e = hwio.GetPinWithMode(actionLedPin, hwio.OUTPUT)
    if e != nil {
        panic(e)
    }
    log.Printf("Set pin %s as actionLed\n",actionLedPin)
}



func readTracker(name string, trackerPin hwio.Pin){

    //value readed from tracker, initially set to 0, because the tracker was innactive
    oldValue := 0

    timeAction := time.Now() // time of the action detected
    timeActionOld := time.Now() // time of the action-1 detected
    // loop
    for theAcq.getState() != stateSTOPPED {
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
	    if  theAcq.getState() != statePAUSED {
                log.Println(dataString)
                theAcq.outputFile.WriteString(dataString)
            }
            oldValue = value

            // Write the value to the led indicating somewhat is happened
	    if  theAcq.getState() != statePAUSED {
                if (value == 1) {
                    hwio.DigitalWrite(actionLed, hwio.HIGH)
                } else {
                    hwio.DigitalWrite(actionLed, hwio.LOW)
                }
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


//WWWWWWWWWWWWW
// http section
//WWWWWWWWWWWWW

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
    theAcq.setState(stateSTOPPED)
    p := &Page{Title: "Index", Body: "Make an action. State:  "+theAcq.state}
    renderTemplate(w,"index",p)
}

func newHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.setState(stateNEW)
    theAcq.createOutputFile()

    // create a new output file
    //var e error
    //theAcq.outputFile, e = os.Create(theAcq.outputFileName)
    //if e != nil {
    //    panic(e)
   // }

    p := &Page{Title: "Index", Body: "New acquisition ready. Select Start to begin it. State: " + theAcq.state}

    renderTemplate(w,"index",p)
}

func startHandler(w http.ResponseWriter, r *http.Request) {

    // manage file depending the previous state 
    if theAcq.getState()== stateSTOPPED {
        theAcq.reopenOutputFile() 
        log.Printf("Reopen file %s\n", theAcq.outputFile);
    }
    theAcq.setState(stateRUNNING)


    //waitTillButtonPushed(buttonA)
    p := &Page{Title: "Start", Body: "State: "+theAcq.state}
    hwio.DigitalWrite(theOshi.statusLed, hwio.HIGH)
    t0 = time.Now()
    log.Println("Beginning.....");

    renderTemplate(w,"start",p)

    // launch the trackers

    log.Printf("There are %v goroutines", runtime.NumGoroutine())
    log.Printf("Launching the Gourutines")
    go readTracker("A", theOshi.trackerA)
    log.Println("Started Tracker A")
    go readTracker("B", theOshi.trackerB)
    log.Println("Started Tracker B")
    go readTracker("C", theOshi.trackerC)
    log.Println("Started Tracker C")
    go readTracker("D", theOshi.trackerD)
    log.Println("Started Tracker D")

    log.Printf("There are %v goroutines", runtime.NumGoroutine())
}

func pauseHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.setState(statePAUSED)

    p := &Page{Title: "Start", Body: "State: "+theAcq.state}
    hwio.DigitalWrite(theOshi.statusLed, hwio.HIGH)

    renderTemplate(w,"start",p)
}

func resumeHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.setState(stateRUNNING)

    p := &Page{Title: "Start", Body: "State: "+theAcq.state}
    hwio.DigitalWrite(theOshi.statusLed, hwio.HIGH)

    renderTemplate(w,"start",p)
}

func stopHandler(w http.ResponseWriter, r *http.Request) {

    theAcq.setState(stateSTOPPED)
    hwio.DigitalWrite(theOshi.statusLed, hwio.LOW)
    log.Println("Finnishing.....");
    // close the GPIO pins
    //hwio.CloseAll()
    theAcq.closeOutputFile() //close the file when finished
    p := &Page{Title: "Stop", Body:"State: "+theAcq.state}
    renderTemplate(w,"stop",p)
    log.Printf("There are %v goroutines", runtime.NumGoroutine())
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
    file_requested := "./"+r.URL.Path
    http.ServeFile(w, r, file_requested)
}

func getRequest(w http.ResponseWriter, r *http.Request){
    file_requested := "./"+r.URL.Path
    http.ServeFile(w, r, file_requested)
}

func main() {

    // setup 
    mux := http.NewServeMux()
    //mux.Handle("/",http.FileServer(http.Dir("data")))
    mux.HandleFunc("/", indexHandler)
    mux.HandleFunc("/index", indexHandler)
    mux.HandleFunc("/new", newHandler)
    mux.HandleFunc("/start", startHandler)
    mux.HandleFunc("/pause", pauseHandler)
    mux.HandleFunc("/resume", resumeHandler)
    mux.HandleFunc("/stop", stopHandler)
    //mux.HandleFunc("/download",getRequest)
    mux.HandleFunc("/"+dataPath+dataFileName, getRequest) // /data/output.csv

    theAcq.initiate();
    theOshi.initiate();

    // starting the web service...
   // http.Handle("/data", http.FileServer(http.Dir("./data")))
    log.Println("Listennig on http://localhost:8080/")
    log.Fatal(http.ListenAndServe(":8080", mux))

    log.Println("Closed http://localhost:8080/")
    // close the GPIO pins
    hwio.CloseAll()
} 
