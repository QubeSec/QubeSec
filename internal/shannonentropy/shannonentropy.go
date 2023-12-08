package shannonentropy

import (
	"fmt"
	"math"
)

func ShannonEntropy(data []byte) float64 {
	// Step 1: Convert bytes array to binary string
	binaryString := bytesToBinary(data)

	// Step 2: Calculate entropy
	return calculateEntropy(binaryString)
}

func bytesToBinary(data []byte) string {
	binaryString := ""
	for _, b := range data {
		// Convert each byte to 8-bit binary representation
		binaryString += fmt.Sprintf("%08b", b)
	}
	return binaryString
}

func calculateEntropy(binaryString string) float64 {
	charCount := make(map[byte]int)
	totalChars := 0

	// Count occurrences of each character
	for i := 0; i < len(binaryString); i++ {
		charCount[binaryString[i]]++
		totalChars++
	}

	// Calculate entropy using Shannon's formula
	entropy := 0.0
	for _, count := range charCount {
		probability := float64(count) / float64(totalChars)
		entropy -= probability * math.Log2(probability)
	}

	return entropy
}
