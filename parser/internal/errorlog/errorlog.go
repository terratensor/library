package errorlog

import (
	"fmt"
	"log"
	"os"
	"time"
)

func Save(errors []string) {
	if len(errors) > 0 {
		// Создание файла для записи ошибок при обработке
		currentTime := time.Now()
		logfile := fmt.Sprintf("./%v_error_log.txt", currentTime.Format("15-04-05_02012006"))

		f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		for _, errorFile := range errors {
			data := []byte(fmt.Sprint(errorFile))
			f.Write(data)
		}
	}
}
