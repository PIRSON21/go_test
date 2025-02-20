package logging

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

func Init() {
	// Открытие файла для логирования
	logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Ошибка при открытии файла для логирования: ", err)
	}

	// Инициализация логгеров
	infoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Логирование успешного старта
	infoLogger.Println("Запуск приложения")
}

func Info(msg string) {
	infoLogger.Println(msg)
}

func Error(msg string, err error) {
	errorLogger.Printf(msg+": %v", err)
}

func Fatal(msg string, err error) {
	errorLogger.Printf(msg+": %v", err)
	os.Exit(1)
}
