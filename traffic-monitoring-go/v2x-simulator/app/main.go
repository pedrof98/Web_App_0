package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

var (
	dsrcPort = flag.Int("dsrc-port", 5001, "Port to send DSRC messages to")
	cv2xPort = flag.Int("cv2x-port", 5002, "Port to send C-V2X messages to")
	host     = flag.String("host", "localhost", "Host to send messages to")
	interval = flag.Int("interval", 200, "Interval between messages in milliseconds")
)

// VehicleInfo represents a simulated vehicle
type VehicleInfo struct {
	ID        uint32
	Latitude  float64
	Longitude float64
	Speed     float32
	Heading   float32
}

func main() {
	// First read environment variables (these take precedence over flags)
	if envHost := os.Getenv("HOST"); envHost != "" {
		*host = envHost
	}
	if envDsrcPort := os.Getenv("DSRC_PORT"); envDsrcPort != "" {
		if p, err := strconv.Atoi(envDsrcPort); err == nil {
			*dsrcPort = p
		}
	}
	if envCv2xPort := os.Getenv("CV2X_PORT"); envCv2xPort != "" {
		if p, err := strconv.Atoi(envCv2xPort); err == nil {
			*cv2xPort = p
		}
	}
	if envInterval := os.Getenv("INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			*interval = i
		}
	}

	// Then parse command line flags (they can override environment variables)
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	log.Printf("Starting V2X simulator sending to %s (DSRC:%d, C-V2X:%d)", *host, *dsrcPort, *cv2xPort)

	// Create vehicles
	vehicles := make([]VehicleInfo, 10)
	for i := range vehicles {
		vehicles[i] = VehicleInfo{
			ID:        uint32(rand.Intn(0xFFFFFF)),
			Latitude:  37.7749 + rand.Float64()*0.1,
			Longitude: -122.4194 + rand.Float64()*0.1,
			Speed:     float32(10 + rand.Intn(30)),
			Heading:   float32(rand.Intn(360)),
		}
	}

	// Create UDP connections
	dsrcAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *host, *dsrcPort))
	if err != nil {
		log.Fatalf("Error resolving DSRC address: %v", err)
	}
	dsrcConn, err := net.DialUDP("udp", nil, dsrcAddr)
	if err != nil {
		log.Fatalf("Error connecting to DSRC port: %v", err)
	}
	defer dsrcConn.Close()

	cv2xAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *host, *cv2xPort))
	if err != nil {
		log.Fatalf("Error resolving C-V2X address: %v", err)
	}
	cv2xConn, err := net.DialUDP("udp", nil, cv2xAddr)
	if err != nil {
		log.Fatalf("Error connecting to C-V2X port: %v", err)
	}
	defer cv2xConn.Close()

	// Send messages at the specified interval
	ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
	defer ticker.Stop()

	msgCount := uint8(0)
	for range ticker.C {
		// Update vehicle positions
		for i := range vehicles {
			// Simulate movement
			speedMetersPerSec := vehicles[i].Speed
			//headingRad := float64(vehicles[i].Heading) * (3.14159 / 180.0)
			
			// Move about 100ms worth of distance (very simplified)
			latChange := float64(speedMetersPerSec) * 0.0000009 * 0.1 * float64(rand.Float32()*0.5+0.75) * float64(rand.Intn(2)*2-1)
			lonChange := float64(speedMetersPerSec) * 0.0000009 * 0.1 * float64(rand.Float32()*0.5+0.75) * float64(rand.Intn(2)*2-1)
			
			vehicles[i].Latitude += latChange
			vehicles[i].Longitude += lonChange
			
			// Randomly change speed and heading
			vehicles[i].Speed += float32(rand.Intn(3) - 1)
			if vehicles[i].Speed < 5 {
				vehicles[i].Speed = 5
			} else if vehicles[i].Speed > 40 {
				vehicles[i].Speed = 40
			}
			
			vehicles[i].Heading += float32(rand.Intn(11) - 5)
			if vehicles[i].Heading < 0 {
				vehicles[i].Heading += 360
			} else if vehicles[i].Heading >= 360 {
				vehicles[i].Heading -= 360
			}
		}

		// Send DSRC BSM
		vehicle := vehicles[rand.Intn(len(vehicles))]
		bsmData := createBSM(vehicle, msgCount)
		_, err = dsrcConn.Write(bsmData)
		if err != nil {
			log.Printf("Error sending DSRC BSM: %v", err)
		} else {
			log.Printf("Sent DSRC BSM from vehicle %08X", vehicle.ID)
		}

		// Send C-V2X BSM (different format)
		vehicle = vehicles[rand.Intn(len(vehicles))]
		cv2xData := createCV2XBSM(vehicle, msgCount)
		_, err = cv2xConn.Write(cv2xData)
		if err != nil {
			log.Printf("Error sending C-V2X BSM: %v", err)
		} else {
			log.Printf("Sent C-V2X BSM from vehicle %08X", vehicle.ID)
		}

		// Occasionally send other message types
		if rand.Intn(10) == 0 {
			// Send a DSRC SPAT message
			spatData := createSPAT(msgCount)
			_, err = dsrcConn.Write(spatData)
			if err != nil {
				log.Printf("Error sending DSRC SPAT: %v", err)
			} else {
				log.Printf("Sent DSRC SPAT message")
			}
		}

		if rand.Intn(20) == 0 {
			// Send a C-V2X DENM message
			denmData := createDENM(msgCount)
			_, err = cv2xConn.Write(denmData)
			if err != nil {
				log.Printf("Error sending C-V2X DENM: %v", err)
			} else {
				log.Printf("Sent C-V2X DENM message")
			}
		}

		msgCount++
	}
}

// createBSM creates a simulated DSRC Basic Safety Message
func createBSM(vehicle VehicleInfo, msgCount uint8) []byte {
	// This is a simplified BSM format for simulation
	buf := new(bytes.Buffer)
	
	// Message type (20 for BSM in J2735)
	buf.WriteByte(20)
	
	// Message content
	binary.Write(buf, binary.BigEndian, vehicle.ID)
	buf.WriteByte(msgCount)
	
	// Timestamp - milliseconds of the minute (0-59999)
	now := time.Now()
	dsec := uint16((now.Second() * 1000) + (now.Nanosecond() / 1000000))
	binary.Write(buf, binary.BigEndian, dsec)
	
	// Position
	lat := int32(vehicle.Latitude * 10000000)
	lon := int32(vehicle.Longitude * 10000000)
	binary.Write(buf, binary.BigEndian, lat)
	binary.Write(buf, binary.BigEndian, lon)
	
	// Elevation (0 for simplicity)
	binary.Write(buf, binary.BigEndian, int32(0))
	
	// Speed in 0.02 m/s units
	speed := uint16(vehicle.Speed * 50)
	binary.Write(buf, binary.BigEndian, speed)
	
	// Heading in 0.0125 degree units
	heading := uint16(vehicle.Heading * 80)
	binary.Write(buf, binary.BigEndian, heading)
	
	// Add some padding for simulated data
	buf.Write(make([]byte, 20))
	
	return buf.Bytes()
}

// createCV2XBSM creates a simulated C-V2X Basic Safety Message
func createCV2XBSM(vehicle VehicleInfo, msgCount uint8) []byte {
	buf := new(bytes.Buffer)
	
	// Message type (1 for C-V2X BSM in our simulation)
	buf.WriteByte(1)
	
	// Interface type (PC5=0, Uu=128)
	interfaceType := byte(0)
	if rand.Intn(10) < 2 { // 20% chance of using network
		interfaceType = 128
	}
	buf.WriteByte(interfaceType)
	
	// The rest is similar to DSRC BSM but with different format
	binary.Write(buf, binary.BigEndian, vehicle.ID)
	buf.WriteByte(msgCount)
	
	// Timestamp
	now := time.Now()
	timestamp := uint32(now.Unix())
	binary.Write(buf, binary.BigEndian, timestamp)
	
	// Position
	binary.Write(buf, binary.BigEndian, vehicle.Latitude)
	binary.Write(buf, binary.BigEndian, vehicle.Longitude)
	
	// Speed in m/s
	binary.Write(buf, binary.BigEndian, vehicle.Speed)
	
	// Heading in degrees
	binary.Write(buf, binary.BigEndian, vehicle.Heading)
	
	// Add some C-V2X specific fields
	qosInfo := byte(rand.Intn(8))
	buf.WriteByte(qosInfo)
	
	// Some additional padding
	buf.Write(make([]byte, 10))
	
	return buf.Bytes()
}

// createSPAT creates a simulated Signal Phase and Timing message
func createSPAT(msgCount uint8) []byte {
	buf := new(bytes.Buffer)
	
	// Message type (13 for SPAT in J2735)
	buf.WriteByte(13)
	
	// Message content
	intersectionID := uint32(100 + rand.Intn(10))
	binary.Write(buf, binary.BigEndian, intersectionID)
	buf.WriteByte(msgCount)
	
	// Number of phases
	phaseCount := byte(4)
	buf.WriteByte(phaseCount)
	
	// Each phase
	for i := byte(0); i < phaseCount; i++ {
		phaseID := byte(i + 1)
		buf.WriteByte(phaseID)
		
		// Light state (0=red, 1=yellow, 2=green)
		lightState := byte(rand.Intn(3))
		buf.WriteByte(lightState)
		
		// Timing info
		startTime := uint16(rand.Intn(6000))
		minEndTime := startTime + uint16(rand.Intn(3000))
		maxEndTime := minEndTime + uint16(rand.Intn(1000))
		
		binary.Write(buf, binary.BigEndian, startTime)
		binary.Write(buf, binary.BigEndian, minEndTime)
		binary.Write(buf, binary.BigEndian, maxEndTime)
	}
	
	return buf.Bytes()
}

// createDENM creates a simulated Decentralized Environmental Notification Message
func createDENM(msgCount uint8) []byte {
	buf := new(bytes.Buffer)
	
	// Message type (3 for DENM in our simulation)
	buf.WriteByte(3)
	
	// Interface type (PC5=0, Uu=128)
	interfaceType := byte(0)
	if rand.Intn(10) < 8 { // 80% chance of using network for alerts
		interfaceType = 128
	}
	buf.WriteByte(interfaceType)
	
	// Message content
	eventID := uint32(rand.Intn(1000000))
	binary.Write(buf, binary.BigEndian, eventID)
	buf.WriteByte(msgCount)
	
	// Event type
	// 1=accident, 2=roadworks, 3=weather, 4=hazard, 5=traffic
	eventType := byte(1 + rand.Intn(5))
	buf.WriteByte(eventType)
	
	// Timestamp
	now := time.Now()
	timestamp := uint32(now.Unix())
	binary.Write(buf, binary.BigEndian, timestamp)
	
	// Position
	latitude := 37.7749 + rand.Float64()*0.1
	longitude := -122.4194 + rand.Float64()*0.1
	binary.Write(buf, binary.BigEndian, latitude)
	binary.Write(buf, binary.BigEndian, longitude)
	
	// Radius
	radius := uint16(100 + rand.Intn(900))
	binary.Write(buf, binary.BigEndian, radius)
	
	// Duration in seconds
	duration := uint16(300 + rand.Intn(3600))
	binary.Write(buf, binary.BigEndian, duration)
	
	// Some additional info
	buf.Write(make([]byte, 20))
	
	return buf.Bytes()
}