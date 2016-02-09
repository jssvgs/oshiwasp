package main

import (
        "log"
	"fmt"
        "github.com/tarm/serial"
	"bytes"
	"encoding/binary"
//	"encoding/csv"
	"bufio"
//	"os"
	"time"

)



type SensorData struct{
	sincroMicroSeconds uint32
	sensorMicroSeconds uint32
	distance uint32
	accX float32
	accY float32
	accZ float32
	gyrX float32
	gyrY float32
	gyrZ float32
}



func main() {

        c := &serial.Config{Name: "/dev/rfcomm1", Baud: 9600}
        s, err := serial.OpenPort(c)
	defer s.Close()

	// activate the readdings sending 'ON'
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
	//log.Println(register)
	//log.Printf(">>>>>>>>>>>>>>")


	var sensorData SensorData

	var sincroMicroSecondsInBytes []byte
	var sensorMicroSecondsInBytes []byte
	var distanceInBytes []byte
	var accXInBytes []byte
	var accYInBytes []byte
	var accZInBytes []byte
	var gyrXInBytes []byte
	var gyrYInBytes []byte
	var gyrZInBytes []byte

	var receptionTime time.Time


	for {
		// clear register and reg
		register = nil
		reg = nil

		//n, err = s.Read(register)
		for len(register) < 38 { // in case of \x24 chars repeted
			reg, err = reader.ReadBytes('\x24');
			if err != nil { log.Fatal(err) }
			register = append(register, reg...)
	 	}
		receptionTime = time.Now()

	//	log.Println(register)
		if (register[0] == '\x23') { // if first byte is '#', lets decode the stream of bytes in register

			//decode the register
	
			sincroMicroSecondsInBytes = register[1:5]
			buf := bytes.NewReader(sincroMicroSecondsInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.sincroMicroSeconds)

			sensorMicroSecondsInBytes = register[5:9]
			buf = bytes.NewReader(sensorMicroSecondsInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.sensorMicroSeconds)

			distanceInBytes = register[9:13]
			buf = bytes.NewReader(distanceInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.distance)

			accXInBytes = register[13:17]
			buf = bytes.NewReader(accXInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.accX)

			accYInBytes = register[17:21]
			buf = bytes.NewReader(accYInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.accY)

			accZInBytes = register[21:25]
			buf = bytes.NewReader(accZInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.accZ)

			gyrXInBytes = register[25:29]
			buf = bytes.NewReader(gyrXInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.gyrX)

			gyrYInBytes = register[29:33]
			buf = bytes.NewReader(gyrYInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.gyrY)

			gyrZInBytes = register[33:37]
			buf = bytes.NewReader(gyrZInBytes)
			err = binary.Read(buf, binary.LittleEndian, &sensorData.gyrZ)

			fmt.Println(receptionTime, sensorData)
		} // if

	}
}

