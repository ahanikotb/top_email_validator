package core

import (
	"encoding/csv"
	"fmt"
	"gorm.io/gorm/clause"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"topemailvalidator/database"
)

func LoadCsvValidOnly(filename string) []map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	var list []map[string]string
	// Create a map to hold the data

	headers := records[0]

	// Loop over the CSV rows and add them to the map
	for _, row := range records[1:] {
		data := make(map[string]string)
		for i, header := range headers {
			data[header] = row[i]
		}
		if data["TOPEMAILSTATUS"] == "valid" {
			list = append(list, data)
		}

	}
	fmt.Println("File Loaded")
	// Print the map
	return list
}

func LoadCSVCatchallOnly(filename string) []map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	var list []map[string]string
	// Create a map to hold the data

	headers := records[0]

	// Loop over the CSV rows and add them to the map
	for _, row := range records[1:] {
		data := make(map[string]string)
		for i, header := range headers {
			data[header] = row[i]
		}
		if data["TOPEMAILSTATUS"] == "catch_all" {
			list = append(list, data)
		}

	}
	fmt.Println("File Loaded")
	// Print the map
	return list
}
func LoadCsv(filename string) []map[string]string {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	var list []map[string]string
	// Create a map to hold the data

	headers := records[0]

	// Loop over the CSV rows and add them to the map
	for _, row := range records[1:] {
		data := make(map[string]string)
		for i, header := range headers {
			data[header] = row[i]
		}
		//data["TOPEMAIL"] = ""
		//data["TOPEMAILSTATUS"] = ""
		list = append(list, data)
	}
	fmt.Println("File Loaded")
	// Print the map
	return list
}
func CountRows(filename string) int {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	return len(records) - 1
}
func GetRowsAndHeaders(filename string) (uint, string) {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Parse the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	return uint(len(records) - 1), strings.Join(records[0], ",")

}

func SaveCsv(fileName string, data []map[string]string, headerString string) {

	// Create a new file for writing
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Create a new CSV writer that writes to the file
	writer := csv.NewWriter(file)

	// Write the header row to the CSV file
	var headers []string
	headers = append(headers, strings.Split(headerString, ",")...)
	headers = append(headers, "TOPEMAIL")
	headers = append(headers, "TOPEMAILSTATUS")
	//for k := range data[0] {
	//	headers = append(headers, k)
	//}
	err = writer.Write(headers)
	if err != nil {
		return
	}

	// Loop through the data and write each row to the CSV file
	for _, row := range data {
		values := make([]string, 0, len(headers))
		for _, header := range headers {
			values = append(values, row[header])
		}
		err := writer.Write(values)
		if err != nil {
			return
		}
	}

	// Flush any remaining data in the buffer to the underlying writer (the file)
	writer.Flush()
}

func processRow(row map[string]string, csvFileSettings database.CsvFileSettings) {
	// Do some processing

	if row[csvFileSettings.EmailColumnName] != "" {
		row["TOPEMAIL"] = row[csvFileSettings.EmailColumnName]
		row["TOPEMAILSTATUS"] = ValidateEmail(row[csvFileSettings.EmailColumnName])
		if row["TOPEMAILSTATUS"] == "failed" && csvFileSettings.BruteForceFailedEmails {
			// if email is invalid and we want to retry it
			email, status := BruteForceValidateEmail(row[csvFileSettings.FirstNameColumnName], row[csvFileSettings.LastNameColumnName], ParseDomain(row[csvFileSettings.WebsiteColumnName]), csvFileSettings.IncludeGenerics)
			if email != "null" {
				row["TOPEMAIL"] = email
				row["TOPEMAILSTATUS"] = status
			}
		}

		return

	}

	if row[csvFileSettings.EmailColumnName] == "" && csvFileSettings.GetMissingEmails {
		if row[csvFileSettings.WebsiteColumnName] == "" {
			return
		}
		email, status := BruteForceValidateEmail(row[csvFileSettings.FirstNameColumnName], row[csvFileSettings.LastNameColumnName], ParseDomain(row[csvFileSettings.WebsiteColumnName]), csvFileSettings.IncludeGenerics)
		if email != "null" {
			row["TOPEMAIL"] = email
			row["TOPEMAILSTATUS"] = status
		}

	}

}

type StatResults struct {
	ValidEmails     uint
	InvalidEmails   uint
	CatchAllEmails  uint
	FullInboxEmails uint
}

func MakeAnalytics(csvFile []map[string]string) StatResults {
	var stats StatResults

	stats.ValidEmails = 0
	stats.InvalidEmails = 0
	stats.CatchAllEmails = 0
	stats.FullInboxEmails = 0

	for i := range csvFile {
		switch csvFile[i]["TOPEMAILSTATUS"] {

		case "valid":
			stats.ValidEmails = stats.ValidEmails + 1
		case "not_valid":
			stats.InvalidEmails = stats.InvalidEmails + 1

		case "catch_all":
			stats.CatchAllEmails = stats.CatchAllEmails + 1
		case "full_inbox":
			stats.FullInboxEmails = stats.FullInboxEmails + 1
		case "failed":
			stats.InvalidEmails = stats.InvalidEmails + 1
		default:
			stats.InvalidEmails = stats.InvalidEmails + 1
		}

	}
	return stats
}

//	func ValidateCSVFile(filepath string, csvFileSettings database.CsvFileSettings) {
//		database.Db := database.OPENDB()
//		var file database.File
//		var verificationStatus database.VerificationStatus
//		database.Db.Preload(clause.Associations).Where("system_file_name = ?", strings.Replace(filepath, "./input/", "", 1)).First(&file)
//
//		database.Db.Where("file_id = ?", file.ID).First(&verificationStatus)
//
//		progressCh := make(chan uint)
//		go func() {
//			for percentage := range progressCh {
//
//				verificationStatus.PercentageDone = uint(percentage)
//				//if percentage == 100 {
//				//	MakeAnalytics(file, verificationStatus)
//				//	verificationStatus.Status = "Done"
//				//}
//				if verificationStatus.Status == "Stopped" {
//					return
//				}
//				database.Db.Save(&verificationStatus)
//			}
//		}()
//		csvFile := LoadCsv(filepath)
//		total := len(csvFile)
//		chunkSize := 100 // number of rows to process per goroutine
//		count := int64(0)
//		var wg sync.WaitGroup
//		for i := 0; i < total; i += chunkSize {
//			end := i + chunkSize
//			if end > total {
//				end = total
//			}
//			wg.Add(1)
//			go func(start, end int) {
//				defer wg.Done()
//				for j := start; j < end; j++ {
//					processRow(csvFile[j], csvFileSettings)
//					atomic.AddInt64(&count, 1)
//					percentage := (float64(count) / float64(total)) * float64(100)
//					if int64(math.Round(percentage))%5 == 0 {
//						progressCh <- uint(math.Round(percentage))
//					}
//				}
//
//			}(i, end)
//		}
//
//		wg.Wait()
//		close(progressCh)
//
//		MakeAnalytics(csvFile, verificationStatus)
//		SaveCsv(strings.Replace(filepath, "input", "output", 1), csvFile)
//		verificationStatus.Status = "Done"
//		database.Db.Save(&verificationStatus)
//	}

func ValidateCatchallCSVFile(filepath string, csvFileSettings database.CsvFileSettings) {

	var file database.File
	var verificationStatus database.VerificationStatus
	database.Db.Preload(clause.Associations).Where("system_file_name = ?", strings.Replace(filepath, "./input/", "", 1)).First(&file)

	database.Db.Where("file_id = ?", file.ID).First(&verificationStatus)

	progressCh := make(chan uint)
	go func() {
		for percentage := range progressCh {

			verificationStatus.PercentageDone = uint(percentage)
			//if percentage == 100 {
			//    MakeAnalytics(file, verificationStatus)
			//    verificationStatus.Status = "Done"
			//}
			if verificationStatus.Status == "Stopped" {
				return
			}
			database.Db.Save(&verificationStatus)
		}
	}()
	csvFile := LoadCSVCatchallOnly(filepath)
	total := len(csvFile)
	chunkSize := 100 // number of rows to process per goroutine
	count := int64(0)
	var wg sync.WaitGroup
	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize
		if end > total {
			end = total
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				processRow(csvFile[j], csvFileSettings)
				atomic.AddInt64(&count, 1)
				percentage := (float64(count) / float64(total)) * float64(100)
				if int64(math.Round(percentage))%5 == 0 {
					progressCh <- uint(math.Round(percentage))
				}
			}

		}(i, end)

	}

	wg.Wait() // Wait for all goroutines to finish before continuing

	close(progressCh)

	SaveCsv(strings.Replace(filepath, "input", "output", 1), csvFile, file.HeaderString)
	stats := MakeAnalytics(csvFile)
	verificationStatus.Status = "Done"
	verificationStatus.InvalidEmails = stats.InvalidEmails
	verificationStatus.ValidEmails = stats.ValidEmails
	verificationStatus.FullInboxEmails = stats.FullInboxEmails
	verificationStatus.CatchAllEmails = stats.CatchAllEmails

	database.Db.Save(&verificationStatus)
}
func ValidateCSVFile(filepath string, csvFileSettings database.CsvFileSettings) {

	var file database.File
	var verificationStatus database.VerificationStatus
	database.Db.Preload(clause.Associations).Where("system_file_name = ?", strings.Replace(filepath, "./input/", "", 1)).First(&file)

	database.Db.Where("file_id = ?", file.ID).First(&verificationStatus)

	progressCh := make(chan uint)
	go func() {
		for percentage := range progressCh {

			verificationStatus.PercentageDone = uint(percentage)
			//if percentage == 100 {
			//    MakeAnalytics(file, verificationStatus)
			//    verificationStatus.Status = "Done"
			//}
			if verificationStatus.Status == "Stopped" {
				return
			}
			database.Db.Save(&verificationStatus)
		}
	}()
	csvFile := LoadCsv(filepath)
	total := len(csvFile)
	chunkSize := 100 // number of rows to process per goroutine
	count := int64(0)
	var wg sync.WaitGroup
	for i := 0; i < total; i += chunkSize {
		end := i + chunkSize
		if end > total {
			end = total
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end; j++ {
				processRow(csvFile[j], csvFileSettings)
				atomic.AddInt64(&count, 1)
				percentage := (float64(count) / float64(total)) * float64(100)
				if int64(math.Round(percentage))%5 == 0 {
					progressCh <- uint(math.Round(percentage))
				}
			}

		}(i, end)

	}

	wg.Wait() // Wait for all goroutines to finish before continuing

	close(progressCh)

	SaveCsv(strings.Replace(filepath, "input", "output", 1), csvFile, file.HeaderString)
	stats := MakeAnalytics(csvFile)
	verificationStatus.Status = "Done"
	verificationStatus.InvalidEmails = stats.InvalidEmails
	verificationStatus.ValidEmails = stats.ValidEmails
	verificationStatus.FullInboxEmails = stats.FullInboxEmails
	verificationStatus.CatchAllEmails = stats.CatchAllEmails

	database.Db.Save(&verificationStatus)
}

//func ValidateCSVFile(filepath string, csvFileSettings database.CsvFileSettings) {
//	var verificationStatus database.VerificationStatus
//	var file database.File
//	database.Db := database.OPENDB()
//	database.Db.Preload(clause.Associations).Where("system_file_name = ?", strings.Replace(filepath, "./input/", "", 1)).First(&file)
//	database.Db.Where("file_id = ?", file.ID).First(&verificationStatus)
//
//	csvFile := LoadCsv(filepath)
//	total := len(csvFile)
//	chunkSize := 100 // number of rows to process per goroutine
//	count := int64(0)
//	var wg sync.WaitGroup
//	for i := 0; i < total; i += chunkSize {
//		end := i + chunkSize
//		if end > total {
//			end = total
//		}
//		wg.Add(1)
//		go func(start, end int) {
//			defer wg.Done()
//			for j := start; j < end; j++ {
//				processRow(csvFile[j], csvFileSettings)
//				atomic.AddInt64(&count, 1)
//				//fmt.Println((float64(count) / float64(total)) * float64(100))
//			}
//		}(i, end)
//	}
//	wg.Wait()
//	SaveCsv(strings.Replace(filepath, "input", "output", 1), csvFile)
//	stats := MakeAnalytics(csvFile)
//	verificationStatus.Status = "Done"
//	verificationStatus.InvalidEmails = stats.InvalidEmails
//	verificationStatus.ValidEmails = stats.ValidEmails
//	verificationStatus.FullInboxEmails = stats.FullInboxEmails
//	verificationStatus.CatchAllEmails = stats.CatchAllEmails
//
//	fmt.Println(verificationStatus)
//	database.Db.Save(&verificationStatus)
//}
