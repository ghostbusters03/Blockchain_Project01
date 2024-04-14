package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Block struct {
	Transactions []string
	PrevHash     string
	Nonce        int
	CurrentHash  string
	MerkleRoot   string
	MerkleTree   [][]string
}

var Blockchain []Block
var BootstrapNodeIP = "localhost" // Bootstrap node IP address
var BootstrapNodePort = 9090      // Bootstrap node port

// Add a global variable to track the bootstrap node
var BootstrapNode = Node{IP: BootstrapNodeIP, Port: BootstrapNodePort}

var mutex sync.Mutex

type Node struct {
	IP   string
	Port int
}

var Nodes []Node

func createBlock(transactions []string, prevHash string) Block {
	merkleRoot, merkleTree := createMerkleRoot(transactions)
	block := Block{
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
		MerkleRoot:   merkleRoot,
		MerkleTree:   merkleTree,
		CurrentHash:  calculateBlockHash(merkleRoot, prevHash, 0),
	}
	return block
}

func calculateBlockHash(merkleRoot, prevHash string, nonce int) string {
	record := prevHash + merkleRoot + fmt.Sprintf("%d", nonce)
	h := sha256.New()
	h.Write([]byte(record))
	return hex.EncodeToString(h.Sum(nil))
}

func createMerkleRoot(transactions []string) (string, [][]string) {
	var levels [][]string
	if len(transactions) == 0 {
		return "", levels
	}
	currentLevel := make([]string, len(transactions))
	for i, tx := range transactions {
		currentLevel[i] = calculateHash(tx)
	}
	levels = append(levels, currentLevel)

	for len(currentLevel) > 1 {
		var nextLevel []string
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				combinedHash := calculateHash(currentLevel[i] + currentLevel[i+1])
				nextLevel = append(nextLevel, combinedHash)
			} else {
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}
		levels = append(levels, nextLevel)
		currentLevel = nextLevel
	}
	return currentLevel[0], levels
}

func mineBlock(block Block, difficulty int) Block {
	target := strings.Repeat("0", difficulty)
	for {
		block.Nonce++
		block.CurrentHash = calculateBlockHash(block.MerkleRoot, block.PrevHash, block.Nonce)
		if strings.HasSuffix(reverseString(block.CurrentHash), target) {
			break
		}
	}
	return block
}

func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func calculateHash(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// Vehicle management functions
func AddVehicle(vin, make, model string, year int, owner string) string {
	transaction := fmt.Sprintf("AddVehicle: %s, %s, %s, %d, Owner: %s", vin, make, model, year, owner)
	fmt.Printf("Vehicle registered: %s %s, Owner: %s\n", make, model, owner)
	return transaction
}

func TransferOwnership(vin, from, to string) string {
	transaction := fmt.Sprintf("TransferOwnership: %s, From: %s, To: %s", vin, from, to)
	fmt.Printf("Ownership of VIN %s transferred from %s to %s\n", vin, from, to)
	return transaction
}

func RecordMaintenance(vin, description string) string {
	transaction := fmt.Sprintf("RecordMaintenance: %s, Description: %s, Date: %s", vin, description, time.Now().Format("2006-01-02"))
	fmt.Printf("Maintenance recorded for %s: %s\n", vin, description)
	return transaction
}

func ReportAccident(vin, description string) string {
	transaction := fmt.Sprintf("ReportAccident: %s, Description: %s, Date: %s", vin, description, time.Now().Format("2006-01-02"))
	fmt.Printf("Accident reported for %s: %s\n", vin, description)
	return transaction
}

func GenerateVehicleHistoryReport(vin string) {
	fmt.Printf("History Report for VIN %s:\n", vin)
	for _, block := range Blockchain {
		for _, tx := range block.Transactions {
			if strings.Contains(tx, vin) {
				fmt.Println(tx)
			}
		}
	}
}

func displayBlocks() {
	for _, block := range Blockchain {
		fmt.Printf("Block - Previous Hash: %s, Merkle Root: %s, Nonce: %d, Current Hash: %s\n", block.PrevHash, block.MerkleRoot, block.Nonce, block.CurrentHash)
		displayMerkleTree(block.MerkleTree)
	}
}

func displayMerkleTree(merkleTree [][]string) {
	if len(merkleTree) == 0 {
		fmt.Println("No Merkle Tree to display.")
		return
	}
	fmt.Println("Merkle Tree:")
	for level, nodes := range merkleTree {
		fmt.Printf("Level %d: ", level)
		for _, node := range nodes {
			fmt.Printf("%s ", node)
		}
		fmt.Println()
	}
	fmt.Println()
}

func registerNode(ip string, port int) {
	mutex.Lock()
	defer mutex.Unlock()
	// Check if the node already exists
	for _, node := range Nodes {
		if node.IP == ip && node.Port == port {
			fmt.Printf("Node %s:%d already registered.\n", ip, port)
			return
		}
	}
	newNode := Node{IP: ip, Port: port}
	Nodes = append(Nodes, newNode)
	fmt.Printf("Node registered: IP: %s, Port: %d\n", ip, port)
	// Add a log statement to check the nodes slice after adding the new node
	fmt.Println("Updated Nodes:", Nodes)
}

// Add a function to send information about known nodes to another node
func sendNodesInfo(conn net.Conn) {
	// Create a string containing information about known nodes
	nodeInfo := ""
	for _, node := range Nodes {
		nodeInfo += fmt.Sprintf("%s:%d\n", node.IP, node.Port)
	}
	// Send the information to the other node
	conn.Write([]byte(nodeInfo))
}

// Modify the handleConnection function to handle messages from other nodes
func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Read messages from the connection
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading from connection:", err.Error())
		}
		return
	}
	message := string(buffer[:n])
	// If the message is requesting node information, send the information
	if message == "getNodesInfo" {
		sendNodesInfo(conn)
	}
}

func connectToBootstrapNode(port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", BootstrapNode.IP, BootstrapNode.Port))
	if err != nil {
		fmt.Println("Error connecting to bootstrap node:", err.Error())
		return
	}
	defer conn.Close()

	// Register the node with the bootstrap node
	registerNode("localhost", port)

	// Request information about existing nodes
	conn.Write([]byte("getNodesInfo"))

	// Receive IP addresses/port numbers of existing nodes from the bootstrap node
	existingNodes := receiveExistingNodes(conn)
	// Connect to existing nodes
	connectedNodes := make(map[string]bool) // Track connected nodes
	for _, node := range existingNodes {
		address := fmt.Sprintf("%s:%d", node.IP, node.Port)
		if !connectedNodes[address] && (node.IP != "localhost" || node.Port != port) {
			// Avoid connecting to itself and skip duplicate connections
			connectToPeer(node)
			connectedNodes[address] = true
		}
	}
	// Add a log statement to check the list of existing nodes after receiving from bootstrap node
	fmt.Println("Existing Nodes after connecting to bootstrap node:", existingNodes)
}

func receiveExistingNodes(conn net.Conn) []Node {
	// Read peer addresses from the connection
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error receiving existing nodes:", err.Error())
		}
		return nil
	}
	peerData := string(buffer[:n])

	// Split the received data to extract individual peer addresses
	existingNodesData := strings.Split(peerData, "\n")
	var existingNodes []Node
	for _, data := range existingNodesData {
		parts := strings.Split(data, ":")
		if len(parts) == 2 {
			port, _ := strconv.Atoi(parts[1])
			node := Node{IP: parts[0], Port: port}
			existingNodes = append(existingNodes, node)
		}
	}
	fmt.Println("Received existing nodes:", existingNodes)
	return existingNodes
}

func connectToPeer(node Node) {
	address := fmt.Sprintf("%s:%d", node.IP, node.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err.Error())
		return
	}
	defer conn.Close()
	fmt.Printf("Connected to peer: %s\n", address)
	// Handle communication with the peer here
}

func displayNetwork() {
	fmt.Println("P2P Network:")

	for _, node := range Nodes {
		fmt.Printf("Node IP: %s, Port: %d\n", node.IP, node.Port)
	}
}

func startServer(port int, isBootstrap bool) {
	// If it's the bootstrap node, register it
	if isBootstrap {
		registerNode("localhost", port)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer ln.Close()
	fmt.Printf("Node listening on port %d\n", port)

	// If it's the bootstrap node, just handle incoming connections
	if isBootstrap {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err.Error())
				continue
			}
			// Handle incoming connections
			go handleConnection(conn)
		}
	}

	if !isBootstrap {
		ticker := time.NewTicker(10 * time.Second) // Adjust the interval as needed
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				// Periodically exchange information about known nodes with other nodes
				for _, node := range Nodes {
					go func(node Node) {
						conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", node.IP, node.Port))
						if err != nil {
							fmt.Println("Error connecting to node:", err.Error())
							return
						}
						defer conn.Close()
						// Send a request for node information
						conn.Write([]byte("getNodesInfo"))
						// Receive information about known nodes from the other node
						existingNodes := receiveExistingNodes(conn)
						// Update the list of known nodes
						mutex.Lock()
						for _, newNode := range existingNodes {
							Nodes = append(Nodes, newNode)
						}
						mutex.Unlock()
					}(node)
				}
			default:
				// Do nothing and continue the loop
			}

			// Accept incoming connections
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err.Error())
				continue
			}
			// Handle incoming connections
			go handleConnection(conn)
		}
	}
}
func atoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println("Error converting string to int:", err.Error())
	}
	return i
}

func main() {
	// Define command-line flag for the port number
	var port int
	flag.IntVar(&port, "port", 8081, "Port number for the node")
	flag.Parse()

	// Start the bootstrap node on a separate goroutine
	go startServer(BootstrapNodePort, true)

	// Connect to the bootstrap node
	connectToBootstrapNode(port)

	// Start the server on the specified port
	go startServer(port, false)

	// Initial transactions and mining
	trans1 := AddVehicle("WDDGF7HB8DA832917", "Mercedes", "C63", 2013, "Murtaza Haider")
	trans2 := TransferOwnership("WDDGF7HB8DA832917", "Murtaza Haider", "Abdullah Tariq")
	trans3 := AddVehicle("WP0CA298X2L001306", "Porsche", "Carerra GT", 2002, "Abdullah Gill")
	trans4 := TransferOwnership("WP0CA298X2L001306", "Abdullah Gill", "Murtaza Haider")
	trans5 := AddVehicle("JT2JA82J1R0019362", "Toyota", "Supra", 1995, "Eesha Shafqat")
	trans6 := TransferOwnership("JT2JA82J1R0019362", "Eesha Shafqat", "Abdullah Gill")
	trans7 := RecordMaintenance("WDDGF7HB8DA832917", "Oil change and brake check")
	trans8 := ReportAccident("WP0CA298X2L001306", "Minor scratch on rear bumper")

	genesisTransactions := []string{trans1, trans2, trans3, trans4, trans5, trans6, trans7, trans8}
	genesisBlock := createBlock(genesisTransactions, "")
	minedGenesisBlock := mineBlock(genesisBlock, 2)
	Blockchain = append(Blockchain, minedGenesisBlock)

	// Display P2P network
	displayNetwork()

	displayBlocks()
	GenerateVehicleHistoryReport("WDDGF7HB8DA832917") // Generate and print history report for a specific VIN

	select {}
}
