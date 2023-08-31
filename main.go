package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"

	vault "github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v3"
)

const kvv2 = "secret"

type Data map[string]interface{}

type Client struct {
	client *vault.Client
}

func newClient() (*Client, error) {
	config := vault.DefaultConfig()
	config.Address = "http://localhost:8200"

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	// authenticate
	client.SetToken("dev-only-token")
	return &Client{
		client: client,
	}, nil
}

func (c *Client) write(mount, path string, data Data) error {
	_, err := c.client.KVv2(mount).Put(context.Background(), path, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) get(mount, path string) (Data, error) {
	secret, err := c.client.KVv2(mount).Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	return secret.Data, nil
}

func (c *Client) version(mount, path string) (int, error) {
	secret, err := c.client.KVv2(mount).Get(context.Background(), path)
	if err != nil {
		return -1, err
	}
	return secret.VersionMetadata.Version, nil
}

// RawYaml is a interface for arbitrary data
type RawYaml map[interface{}]interface{}

// ParsedYaml is a map for the parsed data
// The format will be: map["path/to/secret"]["secret_key"]=interface{}
type ParsedYaml map[string]map[string]interface{}

// Import parses the byte stream into yaml struct.
// The go yaml library is able to handle both yaml and json
func Import(filename string) (parsedYaml ParsedYaml, err error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	parsedYaml = make(ParsedYaml)
	rawYaml := make(RawYaml)
	if err = yaml.Unmarshal(file, &rawYaml); err != nil {
		return nil, err
	}

	parseYaml(rawYaml, &parsedYaml, "")
	return parsedYaml, nil
}

func parseYaml(rawYaml RawYaml, parsedYaml *ParsedYaml, path string) {
	for key, value := range rawYaml {
		// Handle nil values in the yaml data
		if value == nil {
			value = ""
		}

		// Check if the given object is of the same type as the RawYaml data type
		// If true - We know that we have not reached the last element of the structure yet
		if reflect.TypeOf(value).String() == reflect.TypeOf(make(RawYaml)).String() {
			tmpPath := fmt.Sprintf("%s/%s", path, key)
			parseYaml(value.(RawYaml), parsedYaml, tmpPath)
		} else {
			// Check if the key exists in the data structure
			// If it doesn't we create it
			if _, exist := (*parsedYaml)[path]; !exist {
				(*parsedYaml)[path] = make(map[string]interface{})
			}

			// Append the value to the parsed data structure using it's absolute path
			(*parsedYaml)[path][fmt.Sprintf("%v", key)] = value
		}
	}
}

func main() {
	m, err := Import("secrets.yml")
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range m {
		fmt.Println(k)
		fmt.Println(v)
	}
	os.Exit(1)

	client, err := newClient()
	if err != nil {
		log.Fatal(err)
	}

	// write a secret
	path := "foo/bar"
	data := Data{"password": "foo"}

	if err = client.write(kvv2, path, data); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Secret written successfully.")

	// Read a secret in dev mode
	secret, err := client.get(kvv2, path)
	if err != nil {
		log.Fatal(err)
	}

	value, ok := secret["password"].(string)
	if !ok {
		log.Fatalf("value type assertion failed: %T %#v", secret["password"], secret["password"])
	}

	if value != "foo" {
		log.Fatalf("unexpected password value %q retrieved from vault", value)
	}

	ver, err := client.version(kvv2, path)
	if err != nil {
		log.Fatalf("failed to get version: %v", err)
	}
	fmt.Printf("%s\t\t%d", path, ver)
}
