package data

import (
	"io"
	"os"
)

// Lit un fichier et retourne son contenu.
func ReadFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return []byte{}, err
	}

	return content, nil
}

// Écris des données dans un nouveau fichier.
func WriteFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
