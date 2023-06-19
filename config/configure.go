package config

import (
	"gopkg.in/ini.v1"
	"log"
	"strconv"
	"sync"
)

type Config struct {
	cfg  *ini.File
	lock *sync.RWMutex
	data map[string]string
}

var defaultConfig = &Config{lock: new(sync.RWMutex), data: make(map[string]string)}

func DefaultConfig() *Config {
	return defaultConfig
}

func LoadFile(fileName string) (*Config, error) {
	cfg, err := ini.Load(fileName)
	if err == nil {
		var config = &Config{lock: new(sync.RWMutex)}
		config.cfg = cfg
		return config, err
	}
	return nil, err
}
func LoadFiles(fileNames []string) (*Config, error) {
	if len(fileNames) == 1 {
		return LoadFile(fileNames[0])
	}
	cfg, err := ini.Load(fileNames[0], fileNames[1:])
	if err == nil {
		var config = &Config{lock: new(sync.RWMutex)}
		config.cfg = cfg
		return config, err
	}
	return nil, err
}

func (config *Config) SetString(section, key string, value string) {
	config.lock.Lock()
	dataKey := section + key
	config.data[dataKey] = value
	config.lock.Unlock()
}

func (config *Config) getValue(section, key string) string {
	config.lock.RLock()
	defer config.lock.RUnlock()
	dataKey := section + key
	v, ok := config.data[dataKey]
	if ok {
		return v
	}
	if config.cfg.HasSection(section) {
		sectionCfg, err := config.cfg.GetSection(section)
		if err != nil {
			return ""
		}
		if sectionCfg.HasKey(key) {
			k := sectionCfg.Key(key)
			return k.Value()
		}
	}
	return ""

}
func (config *Config) GetString(section, key string) string {
	iSection := config.getValue(section, key)
	return iSection
}
func (config *Config) GetInt(section, key string) int {
	v := config.GetString(section, key)
	if len(v) > 0 {
		iv, err := strconv.Atoi(v)
		if err == nil {
			return iv
		} else {
			log.Panic("格式错误  key:", key)
		}
	}
	return -1
}
func (config *Config) GetBool(section, key string) bool {
	v := config.GetString(section, key)
	if v == "true" {
		return true
	}
	return false
}
func (config *Config) GetBoolOrDefault(section, key string, defaultValue bool) bool {
	v := config.GetString(section, key)
	if len(v) == 0 {
		return defaultValue
	}
	if v == "true" {
		return true
	}
	return false
}

func (config *Config) GetStringOrDefault(section, key string, defaultValue string) string {
	v := config.getValue(section, key)
	if len(v) == 0 {
		return defaultValue
	}
	return v
}
func (config *Config) GetIntOrDefault(section, key string, defaultValue int) int {
	v := config.GetString(section, key)
	if len(v) == 0 {
		return defaultValue
	}
	if len(v) > 0 {
		iv, err := strconv.Atoi(v)
		if err == nil {
			return iv
		} else {
			log.Panic("格式错误  key:", key)
		}
	}
	return defaultValue
}
