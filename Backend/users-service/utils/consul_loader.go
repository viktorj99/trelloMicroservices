package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/consul/api"
)

func LoadCommonPasswordsToConsul(filepath string) error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to connect to Consul: %w", err)
	}

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		password := scanner.Text()
		key := fmt.Sprintf("common_passwords/%s", password)

		_, err := client.KV().Put(&api.KVPair{Key: key, Value: []byte("true")}, nil)
		if err != nil {
			log.Printf("Failed to write key %s to Consul: %v", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fmt.Println("Common passwords loaded into Consul.")
	return nil
}
