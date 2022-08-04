package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	logrus "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type cfg struct {
	DSN         string      `yaml:"DSN"`
	API         string      `yaml:"OSU_API"`
	SessionName string      `yaml:"TMUX_NAME"`
	Sessions    [][3]string `yaml:"TMUX_SESSION"`
	Data        struct {
		Replays     string `yaml:"PATH_REPLAYS"`
		Screenshots string `yaml:"PATH_SCREENSHOT"`
		Avatars     string `yaml:"PATH_AVATAR"`
	}
}

func loadConfig() cfg {
	conf, err := ioutil.ReadFile("config.yaml")
	c := cfg{}

	if err != nil {
		fmt.Println("Could not find config.yaml!")
		os.Exit(1)
	}

	err = yaml.Unmarshal(conf, &c)
	derr := yaml.Unmarshal(conf, &c.Data)

	if err != nil || derr != nil {
		fmt.Println("Invalid config!", err)
		os.Exit(1)
	}

	return c
}

func loadLogger() *logrus.Logger {
	var logger = logrus.New()

	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.TraceLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	return logger
}
