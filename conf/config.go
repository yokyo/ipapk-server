package conf

import (
	"encoding/json"
	"fmt"
	"github.com/phinexdaz/ipapk-server/utils"
	"io/ioutil"
)

var AppConfig *Config

type Config struct {
	Tls		 bool   `json:"tls"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Proxy    string `json:"proxy"`
	Database string `json:"database"`
}

func InitConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &AppConfig); err != nil {
		return err
	}
	// Print out for tip
	AppConfig.Print()

	return nil
}

func (c *Config)Print() error {
	fmt.Printf(" Tls:%v\n Host: %v\n Port: %v\n Proxy: %v\n Database: %v\n",
		c.Tls, c.Host, c.Port, c.Proxy, c.Database)
	return nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%v:%v", c.Host, c.Port)
}

func (c *Config) ProxyURL() string {
	if c.Proxy == "" {
		localIp, err := utils.LocalIP()
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("https://%v:%v", localIp.String(), c.Port)
	}
	return c.Proxy
}
