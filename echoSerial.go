package main

import (
        "log"
//	"fmt"
        "github.com/tarm/serial"
	"bytes"
	"encoding/binary"
	"bufio"
)



func main() {

        c := &serial.Config{Name: "/dev/rfcomm1", Baud: 9600}
        s, err := serial.OpenPort(c)
	defer s.Close()

        _, err = s.Write([]byte("n"))
        if err != nil {
                log.Fatal(err)
       	}

	var register, reg []byte 
	reader := bufio.NewReader(s)
        if err != nil {
                log.Fatal(err)
        }
	// find the begging of an stream of data from the sensors
	register, err = reader.ReadBytes('\x24');
	if err != nil { log.Fatal(err) }
	log.Println(register)
	log.Printf(">>>>>>>>>>>>>>")

/*
        buf := make([]byte,1)
	var lastByte byte = 0x48 // initial value set to 0
      	for itBeggins:=false; !itBeggins; {
       		_, err = s.Read(buf)
        	if err != nil { log.Fatal(err) }

		if (buf[0] == 0x23 || buf[0] == 0x64 ) && lastByte == 0x24 {
			log.Printf(">>>>>>>>>>>>>>")
			itBeggins = true
		}
		lastByte = buf[0]

        	log.Printf("%c", buf[0])
		//fmt.Print(hex.Dump(buf))
		//fmt.Print("<%c", dataFromSerial[0])
		//fmt.Print(dataFromSerial[0])
		//fmt.Println("%c>", dataFromSerial[sizeOfInputLine])
     	 }
*/
	// reads the serial in blocks
	//controlChars := make([]byte,2) // and then 2 of control
	var microSecondsInBytes []byte
	var microSeconds uint32
	var distanceInBytes []byte
	var distance uint32
	var accXInBytes []byte
	var accX float32
	var accYInBytes []byte
	var accY float32
	var accZInBytes []byte
	var accZ float32
	var gyrXInBytes []byte
	var gyrX float32
	var gyrYInBytes []byte
	var gyrY float32
	var gyrZInBytes []byte
	var gyrZ float32


	for {
		// clear register and reg
		register = nil
		reg = nil

		//n, err = s.Read(register)
		for len(register) < 34 { // escape the \x24 char
			reg, err = reader.ReadBytes('\x24');
			if err != nil { log.Fatal(err) }
			register = append(register, reg...)
	 	}
		log.Println(register)
		if (register[0] == '\x23' || register[0] == '\x64') {

		//decode the register

		microSecondsInBytes = register[1:5]
		buf := bytes.NewReader(microSecondsInBytes)
		err = binary.Read(buf, binary.LittleEndian, &microSeconds)
		log.Println("Time: ",microSeconds)		

		distanceInBytes = register[5:9]
		buf = bytes.NewReader(distanceInBytes)
		err = binary.Read(buf, binary.LittleEndian, &distance)
		log.Println("Distance: ",distance)		

		accXInBytes = register[9:13]
		buf = bytes.NewReader(accXInBytes)
		err = binary.Read(buf, binary.LittleEndian, &accX)
		log.Println(accX)		

		accYInBytes = register[13:17]
		buf = bytes.NewReader(accYInBytes)
		err = binary.Read(buf, binary.LittleEndian, &accY)
		log.Println(accY)		

		accZInBytes = register[17:21]
		buf = bytes.NewReader(accZInBytes)
		err = binary.Read(buf, binary.LittleEndian, &accZ)
		log.Println(accZ)		

		gyrXInBytes = register[21:25]
		buf = bytes.NewReader(gyrXInBytes)
		err = binary.Read(buf, binary.LittleEndian, &gyrX)
		log.Println(gyrX)		

		gyrYInBytes = register[25:29]
		buf = bytes.NewReader(gyrYInBytes)
		err = binary.Read(buf, binary.LittleEndian, &gyrY)
		log.Println(gyrY)		

		gyrZInBytes = register[29:33]
		buf = bytes.NewReader(gyrZInBytes)
		err = binary.Read(buf, binary.LittleEndian, &gyrZ)
		log.Println(gyrZ)		


		//_, err = s.Read(controlChars)
		//if err != nil { log.Fatal(err) }
		//log.Println(controlChars[:2])
		//if controlChars[0] != 0x24 {
	//		log.Fatal("not trailing") //not $ al the end
	//	}
	} // if

	}
}

