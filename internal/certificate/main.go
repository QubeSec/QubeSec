package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func main() {

	publicKey, privateKey := createCertificate("rsa:4096", "365", "Post-Quantum")

	fmt.Print(publicKey, privateKey)
}

func createCertificate(algorithm, days, commonName string) (string, string) {

	folderName, err := createFolder()
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return "", ""
	}
	defer cleanUp(folderName)

	cmd := exec.Command(
		"openssl",
		"req",
		"-x509",
		"-new",
		"-newkey",
		algorithm,
		"-keyout",
		folderName+"/tls.key",
		"-out",
		folderName+"/tls.crt",
		"-nodes",
		"-days",
		days,
		"-subj",
		"/CN="+commonName,
	)

	err = cmd.Run()
	if err != nil {
		fmt.Println("Error executing command:", err)
		return "", ""
	}

	certificateFile, err := readFile(folderName + "/tls.crt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return "", ""
	}

	keyFile, err := readFile(folderName + "/tls.key")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return "", ""
	}

	return certificateFile, keyFile
}

func createFolder() (string, error) {

	newUUID := uuid.New().String()

	folderName := "/tmp/qubesec/certificates/" + newUUID

	err := os.MkdirAll(folderName, 0755)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return "", err
	}

	return folderName, nil
}

func cleanUp(folderPath string) error {
	// Get the list of files and subdirectories in the folder
	files, err := filepath.Glob(filepath.Join(folderPath, "*"))
	if err != nil {
		return err
	}

	// Delete files in the folder
	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}

	// Delete the folder itself
	if err := os.Remove(folderPath); err != nil {
		return err
	}

	return nil
}

func readFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
