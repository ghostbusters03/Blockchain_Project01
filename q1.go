package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
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

func main() {
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

	displayBlocks()
	GenerateVehicleHistoryReport("WDDGF7HB8DA832917") // Generate and print history report for a specific VIN
}
