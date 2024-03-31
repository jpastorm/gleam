package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"github.com/jpastorm/gleam/models"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

var ConfigFilePath = "config.sgc"

type Config struct {
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) SaveConfig(token, username, userType string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeDir := usr.HomeDir

	configPath := filepath.Join(homeDir, ConfigFilePath)

	jsonData, err := json.Marshal(models.Config{
		Token:     token,
		Username:  username,
		UserType:  userType,
		FirstTime: false,
	})
	if err != nil {
		return err
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	encryptedHash := make([]byte, aes.BlockSize+len(jsonData))
	iv := encryptedHash[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(encryptedHash[aes.BlockSize:], jsonData)

	err = ioutil.WriteFile(configPath, append(key, encryptedHash...), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) LoadConfig() (models.Config, error) {
	usr, err := user.Current()
	if err != nil {
		return models.Config{}, err
	}
	homeDir := usr.HomeDir

	configPath := filepath.Join(homeDir, ConfigFilePath)

	file, err := os.Open(configPath)
	if err != nil {
		return models.Config{}, err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	cipherText := make([]byte, fileInfo.Size()-32)
	key := make([]byte, 32)
	if _, err := file.Read(key); err != nil {
		return models.Config{}, err
	}
	if _, err := file.Read(cipherText); err != nil {
		return models.Config{}, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return models.Config{}, err
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)

	var cfg models.Config
	if err := json.Unmarshal(cipherText, &cfg); err != nil {
		return models.Config{}, err
	}
	return cfg, nil
}

func (c *Config) DeleteConfigFile() error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeDir := usr.HomeDir

	configPath := filepath.Join(homeDir, ConfigFilePath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(configPath)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) ConfigExists() bool {
	usr, err := user.Current()
	if err != nil {
		return false
	}

	homeDir := usr.HomeDir

	configPath := filepath.Join(homeDir, ConfigFilePath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	return true
}
