package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
)

type Config struct {
	Env         string    `yaml:"env" env-default:"development"`
	Concurrency int       `yaml:"concurrency" env-default:"12"`
	Volume      string    `yaml:"volume" env-default:"./volume"`
	Manticore   Manticore `yaml:"manticore"`
	BatchSize   int       `yaml:"batch_size" env-default:"3000"`
	MinParSize  int       `yaml:"min_par_size" env-default:"300"`
	OptParSize  int       `yaml:"opt_par_size" env-default:"1800"`
	MaxParSize  int       `yaml:"max_par_size" env-default:"3500"`
}

type Manticore struct {
	Engine string `yaml:"engine" env-default:"rowwise"`
	Index  string `yaml:"index" env-default:"library"`
	Host   string `yaml:"host" env-default:"localhost"`
	Port   string `yaml:"port" env-default:"9312"`
}

func MustLoad() *Config {
	// Получаем путь до конфиг-файла из env-переменной CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	// Проверяем существование конфиг-файла
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config

	// Читаем конфиг-файл и заполняем нашу структуру
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}
