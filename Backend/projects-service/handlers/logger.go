package handlers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

var (
	logFile     *os.File
	logger      *log.Logger
	logMutex    sync.Mutex
	currentFile string
	maxLogSize  int64  = 8 * 1024 * 1024 // 8 MB
	logDir      string = "/app/logs"
	logMessage  string
)

type EventSeverity string

const (
	Info    EventSeverity = "INFO"
	Success EventSeverity = "SUCCESS"
	Warning EventSeverity = "WARNING"
	Error   EventSeverity = "ERROR"
)

// mm-yyyy.txt
func getCurrentLogFileName() string {
	now := time.Now()
	return fmt.Sprintf("%s/%02d-%d.txt", logDir, now.Month(), now.Year())
}

func InitLogging() {
	var err error
	currentFile = getCurrentLogFileName()

	logFile, err = os.OpenFile(currentFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
	}

	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	logger.Println("Logging initialized")
}

func LogEvent(eventID string, severity EventSeverity, message string, userID string, projectID string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	// Switch log files if necessary
	if newFile := getCurrentLogFileName(); newFile != currentFile {
		if logFile != nil {
			logFile.Close()
		}
		var err error
		logFile, err = os.OpenFile(newFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Failed to open new log file: %v", err)
		}
		logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
		currentFile = newFile
		logger.Println("Switched to new log file for the new month")
	}

	// Rotation
	if getTotalLogDirSize() > maxLogSize {
		deleteOldestLogFile()
	}

	timestamp := time.Now().Format(time.RFC3339)

	if projectID == "" {
		logMessage = fmt.Sprintf("[%s] EventID: %s | Severity: %s | UserID: %s | Message: %s",
			timestamp, eventID, severity, userID, message)
	} else {
		logMessage = fmt.Sprintf("[%s] EventID: %s | Severity: %s | UserID: %s | ProjectID: %s | Message: %s",
			timestamp, eventID, severity, userID, projectID, message)
	}

	if logger != nil {
		logger.Println(logMessage)
	} else {
		log.Println(logMessage)
	}
}

func getTotalLogDirSize() int64 {
	var totalSize int64
	err := filepath.Walk(logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		log.Printf("Error getting total log directory size: %v", err)
	}
	return totalSize
}

func deleteOldestLogFile() {
	var logFiles []os.FileInfo

	// Read all
	files, err := os.ReadDir(logDir)
	if err != nil {
		log.Printf("Error reading log directory: %v", err)
		return
	}

	// Gathering .txt files
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			info, err := file.Info()
			if err != nil {
				log.Printf("Error getting file info for %s: %v", file.Name(), err)
				continue
			}
			logFiles = append(logFiles, info)
		}
	}

	// Sort files by modification time (oldest first)
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime().Before(logFiles[j].ModTime())
	})

	// Delete oldest
	if len(logFiles) > 0 {
		oldestFile := filepath.Join(logDir, logFiles[0].Name())
		err := os.Remove(oldestFile)
		if err != nil {
			log.Printf("Error deleting oldest log file %s: %v", oldestFile, err)
		} else {
			log.Printf("Deleted oldest log file: %s", oldestFile)
		}
	}
}

func CloseLogging() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		logFile.Close()
	}
}
