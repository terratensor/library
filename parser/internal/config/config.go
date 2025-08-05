package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string    `yaml:"env" env-default:"development"`
	Concurrency    int       `yaml:"concurrency" env-default:"12"`
	MetadataOnly   bool      `yaml:"metadata_only"` // только метаданные, будет обрабатывать только имена файлов и создавать записи в таблицах авторов, категорий и заголовков без полного парсинга содержимого.
	Volume         string    `yaml:"volume" env-default:"./volume"`
	GenresMapPath  string    `yaml:"genres_map_path" env-default:"./config/genres_map.csv"`
	FoldersMapPath string    `yaml:"folders_map_path" env-default:"./config/folders_map.yaml"`
	Manticore      Manticore `yaml:"manticore"`
	BatchSize      int       `yaml:"batch_size" env-default:"3000"`
	MinParSize     int       `yaml:"min_par_size" env-default:"300"`
	OptParSize     int       `yaml:"opt_par_size" env-default:"1800"`
	MaxParSize     int       `yaml:"max_par_size" env-default:"3500"`
	BrokenDocxMode bool      `yaml:"broken_docx_mode" env-default:"false"`
	PDFMode        bool      `yaml:"pdf_mode" env-default:"false"`
	EPUBMode       bool      `yaml:"epub_mode" env-default:"false"`
	Filters        Filters   `yaml:"filters"`
}

type Manticore struct {
	Engine string `yaml:"engine" env-default:"rowwise"`
	Index  string `yaml:"index" env-default:"library"`
	Host   string `yaml:"host" env-default:"localhost"`
	Port   string `yaml:"port" env-default:"9312"`
}

type Filters struct {
	CutBase64          bool `yaml:"cut_base64" env-default:"false"`
	CutBase64Recursive bool `yaml:"cut_base64_recursive" env-default:"false"`
}

func MustLoad() *Config {
	// Получаем путь до конфиг-файла из env-переменной CONFIG_PATH
	configPath := os.Getenv("LIBRARY_CONFIG_PATH")
	if configPath == "" {
		log.Fatal("LIBRARY_CONFIG_PATH environment variable is not set")
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
