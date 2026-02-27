package clog

import (
	"fmt"
	"log"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
	"github.com/yangyin5127/randomhash"
)

var LOG_FILES_PATH string = "./logfiles/"
var CLOG_TXT *os.File

func generateHash() string {
	randomHash := randomhash.New("")
	result, err := randomHash.GenerateHash(8)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not generate random hash. Error %s\n", err)
	}
	return result
}

// enforces rules on what values LOGGER_MAX_TO_ARCHIVE can have in the .env
// generates new logfiles if none are inside the logfiles directory
// makes sure clog.txt can be used for current logging
func Init() {
	godotenv.Load()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	LOGGER_MAX_TO_ARCHIVE := os.Getenv("LOGGER_MAX_TO_ARCHIVE")
	n := 3
	if LOGGER_MAX_TO_ARCHIVE == "" {
		log.Println("LOGGER_MAX_TO_ARCHIVE not set in .env, using default of 3")
	} else {
		k, err := strconv.Atoi(LOGGER_MAX_TO_ARCHIVE)
		if err != nil || n < 3 {
			log.Printf("LOGGER_MAX_TO_ARCHIVE should be ateast 3, using default of 3. Error: %s\n", err)
			n = 3
		} else {
			n = k
		}
	}
	files, err := os.ReadDir(LOG_FILES_PATH)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not fetch logfiles. Error: %s\n", err)
	}
	len := len(files)
	if len == 0 {
		for i := 0; i < n; i++ {
			file, err := os.Create(LOG_FILES_PATH + "log_" + generateHash() + ".txt")
			if err != nil {
				log.Fatal("FATAL ERROR: could not initialize logfiles/. Error: %s\n", err)
			}
			file.Close()
		}
		file, err := os.Create(LOG_FILES_PATH + "clog.txt")
		if err != nil {
			log.Fatalf("FATAL ERROR: could not create clog.txt. Error: %s\n", err)
		}
		file.Close()
	} else {
		resolveClogTxt(files, n)
	}
	CLOG_TXT, err = os.OpenFile(LOG_FILES_PATH + "clog.txt", os.O_RDWR | os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not open clog.txt. Error: %s\n", err)
	}
	log.SetOutput(CLOG_TXT)
}


// returns if clog.txt is available after recovering from any errors caused by an interruption
// of the previous run. Otherwise creates new clog.txt
func resolveClogTxt(logfiles []os.DirEntry, n int) {
	len := len(logfiles)
	for _, file := range logfiles {
		if "clog.txt" == file.Name() {
			if len == n {
				file_, err := os.Create(LOG_FILES_PATH + "log_" + generateHash() + ".txt")
				if err != nil {
					log.Fatalf("FATAL ERROR: could not recover from previous run, creation of logfile failed. Error %s\n", err)
				}
				file_.Close()
			} else if len != n + 1 {
				log.Fatal("clog.txt + other logfiles should add up to LOGGER_MAX_TO_ARCHIVE + 1")
			}
			return
		}
	}
	if len != n {
		log.Fatal("Need LOGGER_MAX_TO_ARCHIVE logfiles before adding clog.txt")
	}
	file, err := os.Create(LOG_FILES_PATH + "clog.txt")
	if err != nil {
		log.Fatal("FATAL ERROR: could not create clog.txt. Error %s\n", err)
	}
	file.Close()
}

func getMaxSize() int64 {
	var n int64
	n_ := os.Getenv("LOGGER_SIZE_B")
	if n_ == "" {
		log.Println("LOGGER_SIZE_B not set, using default of 5MB")
		n = 5000000
	} else {
		k, err := strconv.ParseInt(n_, 10, 64)
		if err != nil || n < 0 {
			log.Printf("LOGGER_SIZE_B not set correctly, using default of 5MB. Error: %s")
			n = 5000000
		} else {
			n = k
		}
	}
	return n
}

func getOldest() string {
	oldest := time.Now()
	oldestFile := ""
	files, err := os.ReadDir(LOG_FILES_PATH)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not read logfiles/. Error: %s\n", err)
	}
	for _, file := range files {
		if "clog.txt" == file.Name() {
			continue
		}
		info, err := file.Info()
		if err != nil {
			log.Fatalf("FATAL ERROR: could not get file info. Error %s\n", err)
		}
		if info.ModTime().Before(oldest) {
			oldest = info.ModTime()
			oldestFile = file.Name()
		}
	}
	return oldestFile
}

func rotate() {
	CLOG_TXT.Close()
	oldest := getOldest()
	err := os.Remove(LOG_FILES_PATH + oldest)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not remove oldest file. Error: %s\n", err)
	}
	err = os.Rename(LOG_FILES_PATH + "clog.txt", LOG_FILES_PATH + oldest)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not rename clog.txt to oldest file. Error: %s\n", err)
	}
	file, err := os.Create(LOG_FILES_PATH + "clog.txt")
	if err != nil {
		log.Fatalf("FATAL ERROR: could not create clog.txt. Error %s\n", err)
	}
	file.Close()
	CLOG_TXT, err = os.OpenFile(LOG_FILES_PATH + "clog.txt", os.O_RDWR | os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("FATAL ERROR: could not open clog.txt. Error %s\n", err)
	}
	log.SetOutput(CLOG_TXT)
}

func Fatal(v ...any) {
	n := getMaxSize()
	stat, err := CLOG_TXT.Stat()
	if err != nil {
		log.Fatalf("FATAL ERROR: could not get the stats for clog.txt. Error: %s\n", err)
	}
	if stat.Size() > n {
		rotate()
	}
	devMode := os.Getenv("DEV_MODE")
	if devMode == "" || devMode == "on" {
		fmt.Println(v...)
	}
	log.Fatal(v...)
}

func Fatalf(format string, v ...any) {
	n := getMaxSize()
	stat, err := CLOG_TXT.Stat()
	if err != nil {
		log.Fatalf("FATAL ERROR: could not get the stats for clog.txt. Error: %s\n", err)
	}
	if stat.Size() > n {
		rotate()
	}
	devMode := os.Getenv("DEV_MODE")
	if devMode == "" || devMode == "on" {
		fmt.Printf(format, v...)
	}
	log.Fatalf(format, v...)
}

func Printf(msg string, args ...any) {
	stat, err := CLOG_TXT.Stat()
	if err != nil {
		log.Fatalf("FATAL ERROR: could not get the stats for clog.txt. Error: %s\n", err)
	}
	n := getMaxSize()
	if stat.Size() > n {
		rotate()
	}
	devMode := os.Getenv("DEV_MODE")
	if devMode == "" || devMode == "on" {
		fmt.Printf(msg, args...)
	}
	log.Printf(msg, args...)
}

func Println(v ...any) {
	stat, err := CLOG_TXT.Stat()
	if err != nil {
		log.Fatalf("FATAL ERROR: could not get the stats for clog.txt. Error: %s\n", err)
	}
	n := getMaxSize()
	if stat.Size() > n {
		rotate()
	}
	devMode := os.Getenv("DEV_MODE")
	if devMode == "" || devMode == "on" {
		fmt.Println(v...)
	}
	log.Println(v...)
}
