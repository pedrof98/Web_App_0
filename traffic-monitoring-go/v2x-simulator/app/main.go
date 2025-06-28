package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"v2x-simulator/app/api"
	"v2x-simulator/app/attacks"
	"v2x-simulator/app/config"
)

var (
	dsrcPort = flag.Int("dsrc-port", 5001, "Port to send DSRC messages to")
	cv2xPort = flag.Int("cv2x-port", 5002, "Port to send C-V2X messages to")
	host     = flag.String("host", "localhost", "Host to send messages to")
	interval = flag.Int("interval", 200, "Interval between messages in milliseconds")
	apiPort  = flag.Int("api-port", 8081, "Port for attack trigger API")
	vehicles []VehicleInfo // move vehicles to package level for API access
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
	// Load configuration (includes attack settings)
	cfg := config.LoadConfig()

	// Environment variables override flags
	if envHost := os.Getenv("HOST"); envHost != "" {
		*host = envHost
	} else if cfg.Host != "" {
		*host = cfg.Host
	}

	if envDsrcPort := os.Getenv("DSRC_PORT"); envDsrcPort != "" {
		if p, err := strconv.Atoi(envDsrcPort); err == nil {
			*dsrcPort = p
		}
	} else {
		*dsrcPort = cfg.DSRCPort
	}

	if envCv2xPort := os.Getenv("CV2X_PORT"); envCv2xPort != "" {
		if p, err := strconv.Atoi(envCv2xPort); err == nil {
			*cv2xPort = p
		}
	} else {
		*cv2xPort = cfg.CV2XPort
	}

	if envInterval := os.Getenv("INTERVAL"); envInterval != "" {
		if i, err := strconv.Atoi(envInterval); err == nil {
			*interval = i
		}
	} else {
		*interval = cfg.Interval
	}

	flag.Parse()
	if envApiPort := os.Getenv("API_PORT"); envApiPort != "" {
		if p, err := strconv.Atoi(envApiPort); err == nil {
			*apiPort = p
		}
	}
	rand.Seed(time.Now().UnixNano())

	log.Printf("V2X Simulator starting: %s (DSRC:%d, C-V2X:%d)", *host, *dsrcPort, *cv2xPort)
	log.Printf("Attack simulation: enabled=%v, rate=%d%%", cfg.AttackEnabled, cfg.AttackRate)

	// Create vehicles
	vehicles = make([]VehicleInfo, cfg.VehicleCount)
	for i := range vehicles {
		vehicles[i] = VehicleInfo{
			ID:        uint32(rand.Intn(0xFFFFFF)),
			Latitude:  37.7749 + rand.Float64()*0.1,
			Longitude: -122.4194 + rand.Float64()*0.1,
			Speed:     float32(10 + rand.Intn(30)),
			Heading:   float32(rand.Intn(360)),
		}
	}

	go startAPIServer()
	go processAttackTriggers()

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
	attackCount := 0

	log.Printf("Starting V2X message simulation with %d vehicles", len(vehicles))

	for range ticker.C {
		// Update vehicle positions
		for i := range vehicles {
			updateVehiclePosition(&vehicles[i])
		}

		// Select vehicle for this message
		vehicle := &vehicles[rand.Intn(len(vehicles))]

		// Determine if this should be an attack scenario
		var attackScenario *attacks.AttackScenario
		if cfg.AttackEnabled && rand.Intn(100) < cfg.AttackRate {
			attackScenario = attacks.GenerateAttackScenario()
			attackScenario.ApplyAttackToVehicle((*attacks.Vehicle)(vehicle))
			attackCount++
			log.Printf("Attack #%d: %s", attackCount, attackScenario.GetAttackDescription())
		}

		// Send DSRC BSM
		bsmData := createBSM(*vehicle, msgCount, attackScenario)
		_, err = dsrcConn.Write(bsmData)
		if err != nil {
			log.Printf("Error sending DSRC BSM: %v", err)
		} else if attackScenario != nil {
			log.Printf("Sent ATTACK DSRC BSM from vehicle %08X [%s]", vehicle.ID, attackScenario.Type)
			api.DatasetExporter.AddAttackMessage(vehicle.ID, vehicle.Latitude, vehicle.Longitude, vehicle.Speed, vehicle.Heading, "DSRC", "BSM", attackScenario)
		} else if msgCount%20 == 0 {
			log.Printf("Sent DSRC BSM from vehicle %08X (msg #%d)", vehicle.ID, msgCount)
		}

		if attackScenario == nil {
			api.DatasetExporter.AddNormalMessage(vehicle.ID, vehicle.Latitude, vehicle.Longitude, vehicle.Speed, vehicle.Heading, "DSRC", "BSM")
		}

		// Send C-V2X BSM (different vehicle, different format)
		vehicle = &vehicles[rand.Intn(len(vehicles))]
		var cv2xAttack *attacks.AttackScenario
		if cfg.AttackEnabled && rand.Intn(100) < cfg.AttackRate {
			cv2xAttack = attacks.GenerateAttackScenario()
			cv2xAttack.ApplyAttackToVehicle((*attacks.Vehicle)(vehicle))
		}

		cv2xData := createCV2XBSM(*vehicle, msgCount, cv2xAttack)
		_, err = cv2xConn.Write(cv2xData)
		if err != nil {
			log.Printf("Error sending C-V2X BSM: %v", err)
		}

		if cv2xAttack != nil {
			api.DatasetExporter.AddAttackMessage(vehicle.ID, vehicle.Latitude, vehicle.Longitude, vehicle.Speed, vehicle.Heading, "CV2X", "BSM", cv2xAttack)
		} else {
			api.DatasetExporter.AddNormalMessage(vehicle.ID, vehicle.Latitude, vehicle.Longitude, vehicle.Speed, vehicle.Heading, "CV2X", "BSM")
		}

		// Occasionally send other message types
		if rand.Intn(10) == 0 {
			spatData := createSPAT(msgCount)
			_, err = dsrcConn.Write(spatData)
			if err != nil {
				log.Printf("Error sending DSRC SPAT: %v", err)
			}
			api.DatasetExporter.AddNormalMessage(0, 0, 0, 0, 0, "DSRC", "SPAT") // SPAT doesn't have a specific vehicle
		}

		if rand.Intn(20) == 0 {
			denmData := createDENM(msgCount)
			_, err = cv2xConn.Write(denmData)
			if err != nil {
				log.Printf("Error sending C-V2X DENM: %v", err)
			}
			api.DatasetExporter.AddNormalMessage(0, 0, 0, 0, 0, "CV2X", "DENM") // Infrastructure message, no vehicle data
		}

		msgCount++
	}
}

// updateVehiclePosition simulates vehicle movement
func updateVehiclePosition(vehicle *VehicleInfo) {
	speedMetersPerSec := vehicle.Speed

	// Move about 200ms worth of distance (very simplified)
	latChange := float64(speedMetersPerSec) * 0.0000009 * 0.2 * float64(rand.Float32()*0.5+0.75) * float64(rand.Intn(2)*2-1)
	lonChange := float64(speedMetersPerSec) * 0.0000009 * 0.2 * float64(rand.Float32()*0.5+0.75) * float64(rand.Intn(2)*2-1)

	vehicle.Latitude += latChange
	vehicle.Longitude += lonChange

	// Randomly change speed and heading
	vehicle.Speed += float32(rand.Intn(3) - 1)
	if vehicle.Speed < 5 {
		vehicle.Speed = 5
	} else if vehicle.Speed > 40 {
		vehicle.Speed = 40
	}

	vehicle.Heading += float32(rand.Intn(11) - 5)
	if vehicle.Heading < 0 {
		vehicle.Heading += 360
	} else if vehicle.Heading >= 360 {
		vehicle.Heading -= 360
	}
}

// createBSM creates a J2735-compliant DSRC Basic Safety Message with optional attack data
func createBSM(vehicle VehicleInfo, msgCount uint8, attack *attacks.AttackScenario) []byte {
	buf := new(bytes.Buffer)

	// Message type (20 for BSM in J2735)
	buf.WriteByte(20)

	// Message content (following J2735 standard)
	binary.Write(buf, binary.BigEndian, vehicle.ID)
	buf.WriteByte(msgCount)

	// Timestamp - milliseconds of the minute (0-59999)
	now := time.Now()
	dsec := uint16((now.Second() * 1000) + (now.Nanosecond() / 1000000))
	binary.Write(buf, binary.BigEndian, dsec)

	// Position (in 1/10 microdegrees as per J2735)
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

	// Add attack-specific data if this is an attack scenario
	if attack != nil {
		// Add attack signature flags
		switch attack.Type {
		case attacks.InvalidSignatureAttack:
			// Add invalid signature marker (special byte sequence)
			buf.WriteByte(0xFF) // Invalid signature flag
			buf.WriteByte(0x00) // Signature error code
		case attacks.PositionJumpAttack:
			// Add position anomaly marker
			buf.WriteByte(0xAA) // Position anomaly flag
			buf.WriteByte(0x01) // Position jump type
		case attacks.SpeedJumpAttack:
			// Add speed anomaly marker
			buf.WriteByte(0xAA) // Speed anomaly flag
			buf.WriteByte(0x02) // Speed jump type
		case attacks.MessageFloodingAttack:
			// Add flooding marker
			buf.WriteByte(0xBB) // Flooding flag
			buf.WriteByte(0x01) // High frequency type
		default:
			// Regular padding
			buf.Write(make([]byte, 2))
		}

		// Add confidence level for attack detection
		if len(attack.Anomalies) > 0 {
			if confidence, ok := attack.Anomalies[0]["confidence"].(float64); ok {
				confidenceByte := uint8(confidence * 255) // Convert 0.0-1.0 to 0-255
				buf.WriteByte(confidenceByte)
			} else {
				buf.WriteByte(200) // Default high confidence
			}
		} else {
			buf.WriteByte(200) // Default high confidence
		}
	} else {
		// Regular padding for normal messages
		buf.Write(make([]byte, 3))
	}

	// Additional padding to maintain consistent message size
	buf.Write(make([]byte, 17))

	return buf.Bytes()
}

// createCV2XBSM creates a C-V2X Basic Safety Message with optional attack data
func createCV2XBSM(vehicle VehicleInfo, msgCount uint8, attack *attacks.AttackScenario) []byte {
	buf := new(bytes.Buffer)

	// Message type (1 for C-V2X BSM)
	buf.WriteByte(1)

	// Interface type (PC5=0, Uu=128)
	interfaceType := byte(0)
	if rand.Intn(10) < 2 {
		interfaceType = 128
	}
	buf.WriteByte(interfaceType)

	// Rest is similar to DSRC BSM but with C-V2X format
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

	// QoS info
	qosInfo := byte(rand.Intn(8))
	buf.WriteByte(qosInfo)

	// Attack-specific markers for C-V2X
	if attack != nil {
		// C-V2X specific attack indicators
		buf.WriteByte(0xCC)                 // C-V2X attack marker
		buf.WriteByte(byte(attack.Type[0])) // First char of attack type
	} else {
		buf.Write(make([]byte, 2))
	}

	// Additional padding
	buf.Write(make([]byte, 8))

	return buf.Bytes()
}

// createSPAT creates a J2735-compliant Signal Phase and Timing message
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

	// Each phase (following J2735 format)
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

// createDENM creates a C-V2X Decentralized Environmental Notification Message
func createDENM(msgCount uint8) []byte {
	buf := new(bytes.Buffer)

	// Message type (3 for DENM)
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

	// Radius and duration
	radius := uint16(100 + rand.Intn(900))
	duration := uint16(300 + rand.Intn(3600))
	binary.Write(buf, binary.BigEndian, radius)
	binary.Write(buf, binary.BigEndian, duration)

	// Additional info
	buf.Write(make([]byte, 20))

	return buf.Bytes()
}

// startAPIServer starts the HTTP API for attack triggers
func startAPIServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/attack/trigger", corsHandler(api.HandleAttackTrigger))
	mux.HandleFunc("/attack/status", corsHandler(api.HandleAttackStatus))
	mux.HandleFunc("/health", corsHandler(healthHandler))
	mux.HandleFunc("/dataset/start", corsHandler(api.HandleDatasetStart))
	mux.HandleFunc("/dataset/stop", corsHandler(api.HandleDatasetStop))
	mux.HandleFunc("/dataset/export", corsHandler(api.HandleDatasetExport))
	mux.HandleFunc("/dataset/status", corsHandler(api.HandleDatasetStatus))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *apiPort),
		Handler: mux,
	}

	log.Printf("Attack trigger API listening on port %d", *apiPort)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Error starting API server: %v", err)
	}
}

// corsHandler adds CORS headers
func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// healthHandler handles health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":             "healthy",
		"simulation_running": true,
		"vehicle_count":      len(vehicles),
		"timestamp":          time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processAttackTriggers processes on-demand attack requests
func processAttackTriggers() {
	for req := range api.AttackTriggerChan {
		log.Printf("Processing on-demand attack: %s (count: %d)", req.AttackType, req.Count)

		// Select target vehicle
		var targetVehicle VehicleInfo
		if req.VehicleID != "" {
			found := false
			for _, v := range vehicles {
				if fmt.Sprintf("VEH-%08X", v.ID) == req.VehicleID {
					targetVehicle = v
					found = true
					break
				}
			}
			if !found {
				targetVehicle = vehicles[rand.Intn(len(vehicles))]
			}
		} else {
			targetVehicle = vehicles[rand.Intn(len(vehicles))]
		}

		// Generate attacks
		attackScenario := api.GetAttackScenario(req.AttackType)
		for i := 0; i < req.Count; i++ {
			attackVehicle := targetVehicle
			attackScenario.ApplyAttackToVehicle((*attacks.Vehicle)(&attackVehicle))

			generateAttackMessage(attackVehicle, attackScenario)

			if req.AttackType == "message_flooding" && i < req.Count-1 {
				time.Sleep(time.Millisecond * 50)
			}
		}

		log.Printf("Completed on-demand attack: %s (%d events)", req.AttackType, req.Count)
	}
}

// generateAttackMessage generates both DSRC and C-V2X attack messages
func generateAttackMessage(vehicle VehicleInfo, attack *attacks.AttackScenario) {
	// Generate DSRC message with attack
	dsrcConn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *host, *dsrcPort))
	if err == nil {
		message := createAttackMessage(vehicle, attack, "DSRC")
		dsrcConn.Write(message)
		dsrcConn.Close()

		// Record attack message in dataset
		api.DatasetExporter.AddAttackMessage(vehicle.ID, vehicle.Latitude, vehicle.Longitude, vehicle.Speed, vehicle.Heading, "DSRC", "BSM", attack)
	}

	// Generate C-V2X message with attack
	cv2xConn, err := net.Dial("udp", fmt.Sprintf("%s:%d", *host, *cv2xPort))
	if err == nil {
		message := createAttackMessage(vehicle, attack, "CV2X")
		cv2xConn.Write(message)
		cv2xConn.Close()

		// Record attack message in dataset
		api.DatasetExporter.AddAttackMessage(vehicle.ID, vehicle.Latitude, vehicle.Longitude, vehicle.Speed, vehicle.Heading, "CV2X", "BSM", attack)
	}
}

// createAttackMessage creates binary message with attack markers
func createAttackMessage(vehicle VehicleInfo, attack *attacks.AttackScenario, protocol string) []byte {
	var buf bytes.Buffer

	// Protocol header
	if protocol == "DSRC" {
		buf.WriteByte(0x00) // DSRC identifier
		buf.WriteByte(0x14) // BSM ID
	} else {
		buf.WriteByte(0xC2) // C-V2X identifier
		buf.WriteByte(0x55) // Version
	}

	// Vehicle data
	binary.Write(&buf, binary.BigEndian, vehicle.ID)
	binary.Write(&buf, binary.BigEndian, uint32(time.Now().UnixMilli()%60000))
	binary.Write(&buf, binary.BigEndian, int32(vehicle.Latitude*10000000))
	binary.Write(&buf, binary.BigEndian, int32(vehicle.Longitude*10000000))
	binary.Write(&buf, binary.BigEndian, uint16(vehicle.Speed*50))
	binary.Write(&buf, binary.BigEndian, uint16(vehicle.Heading*80))

	// Attack marker
	buf.WriteByte(0xAA) // Attack marker
	buf.WriteByte(0xFF) // Attack indicator

	// Attack type byte
	attackByte := byte(0x00)
	switch attack.Type {
	case attacks.PositionJumpAttack:
		attackByte = 0x01
	case attacks.SpeedJumpAttack:
		attackByte = 0x02
	case attacks.InvalidSignatureAttack:
		attackByte = 0x03
	case attacks.MessageFloodingAttack:
		attackByte = 0x04
	case attacks.ReplayAttack:
		attackByte = 0x05
	case attacks.TrustLevelAttack:
		attackByte = 0x06
	}
	buf.WriteByte(attackByte)

	return buf.Bytes()
}
