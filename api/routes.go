package api

import (
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"topemailvalidator/core"
	"topemailvalidator/database"
)

type EmailFindRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Domain    string `json:"domain"`
}
type ValidateEmailRequest struct {
	Email string `json:"email"`
}

const secretKEY = "12312412414124"

func MakeRoutes(r *gin.Engine) {
	r.POST("/api/upload", requireAuth, UploadFile)
	r.GET("/api/file/:id/download", requireAuth, DownloadFile)
	r.GET("/api/file/:id/download_valid", requireAuth, DownloadFileValidOnly)
	r.GET("/api/validate_email", ValidateEmail)
	r.GET("/api/find_email", FindEmail)
	r.GET("/api/get_files", requireAuth, GetFiles)
	r.GET("/api/list/:id/start_verification", requireAuth, StartVerification)
	r.GET("/api/list/:id/progress", requireAuth, GetVerificationProgress)
	r.GET("/api/list/:id/start_catchall_verification", requireAuth, StartCatchallVerification)
	AddAuthRoutes(r)
}

func StartCatchallVerification(c *gin.Context) {
	fileID := c.Param("id")

	var file database.File
	fId, _ := strconv.Atoi(fileID)
	file.ID = uint(fId)
	var veriStat database.VerificationStatus
	var settings database.CsvFileSettings

	database.Db.Preload(clause.Associations).First(&file)
	database.Db.Where("file_id = ?", fileID).First(&veriStat)
	database.Db.Where("file_id = ?", fileID).First(&settings)

	go func() {
		fmt.Println("Starting verification")
		fmt.Println("./input/" + file.SystemFileName)
		go core.ValidateCatchallCSVFile("./input/"+file.SystemFileName, settings)
		veriStat.Status = "Verifying_Catchall"
		//ProcessID

		database.Db.Save(&veriStat)
	}()

	//database.Db.First(&file)
	//fmt.Println(file)

}

func GetVerificationProgress(c *gin.Context) {
	fileID := c.Param("id")
	// Check if file ID is valid
	if _, err := strconv.Atoi(fileID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file ID",
		})
		return
	}

	// Retrieve file information from database
	var verificationSatuts database.VerificationStatus

	database.Db.Where("file_id = ?", fileID).First(&verificationSatuts)

	//// Check if file exists
	//if file.ID == 0 {
	//	c.JSON(http.StatusNotFound, gin.H{
	//		"error": "File not found",
	//	})
	//	return
	//}

	c.JSON(http.StatusOK, gin.H{
		"progress": verificationSatuts,
	})
}
func GetFiles(c *gin.Context) {
	userID, _ := c.Get("userID")

	//database.Db := database.OPENDB()
	var files []database.File
	database.Db.Where("user_id = ?", userID).Find(&files)
	for i := range files {
		var verificationStatus database.VerificationStatus
		database.Db.Where("file_id = ?", files[i].ID).First(&verificationStatus)
		files[i].VerificationStatus = verificationStatus
	}
	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
	return
}

func FindEmail(c *gin.Context) {
	var requestBody EmailFindRequest
	err := c.BindJSON(&requestBody)
	if err != nil {
		fmt.Println(err.Error())
	}
	res, status := core.BruteForceValidateEmail(requestBody.FirstName, requestBody.LastName, core.ParseDomain(requestBody.Domain), false)
	c.JSON(http.StatusOK, gin.H{
		"email":  res,
		"status": status,
	})

}
func ValidateEmail(c *gin.Context) {
	var requestBody ValidateEmailRequest
	err := c.BindJSON(&requestBody)
	if err != nil {
		fmt.Println(err.Error())
	}
	res := core.ValidateEmail(requestBody.Email)
	c.JSON(http.StatusOK, gin.H{
		"email":  requestBody.Email,
		"status": res,
	})
	return
}

func DownloadFileValidOnly(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	// Get the file path from the query parameter
	fileID := c.Param("id")

	var file database.File

	fId, _ := strconv.Atoi(fileID)
	file.ID = uint(fId)

	database.Db.First(&file)
	//
	fileName := filepath.Base(file.SystemFileName)
	//
	//c.Header("Content-Disposition", "attachment; filename="+fileName)
	//// Return the file to the client
	//c.File("./output/" + file.SystemFileName)

	csvData := core.LoadCsvValidOnly("./output/" + file.SystemFileName)
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	SendCSV(c, csvData, file.HeaderString)
}
func SendCSV(c *gin.Context, fileSlice []map[string]string, headerString string) {
	csvWriter := csv.NewWriter(c.Writer)
	defer csvWriter.Flush()

	var header []string
	header = append(header, strings.Split(headerString, ",")...)
	header = append(header, "TOPEMAIL")
	header = append(header, "TOPEMAILSTATUS")
	//header := []string{}
	//for key := range fileSlice[0] {
	//	header = append(header, key)
	//}
	csvWriter.Write(header)

	for _, row := range fileSlice {
		values := make([]string, 0, len(header))
		for _, header := range header {
			values = append(values, row[header])
		}
		err := csvWriter.Write(values)
		if err != nil {
			return
		}
	}

	//for _, row := range fileSlice {
	//	values := []string{}
	//	for _, value := range row {
	//		values = append(values, value)
	//	}
	//	csvWriter.Write(values)
	//}

	c.Header("Content-Type", "text/csv")
}

func DownloadFile(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	// Get the file path from the query parameter
	fileID := c.Param("id")

	var file database.File

	fId, _ := strconv.Atoi(fileID)
	file.ID = uint(fId)

	database.Db.First(&file)

	fileName := filepath.Base(file.SystemFileName)

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	// Return the file to the client
	c.File("./output/" + file.SystemFileName)
}

func StartVerification(c *gin.Context) {
	//userID, _ := c.Get("userID")
	fileID := c.Param("id")

	var file database.File
	fId, _ := strconv.Atoi(fileID)
	file.ID = uint(fId)
	var veriStat database.VerificationStatus
	var settings database.CsvFileSettings

	database.Db.Preload(clause.Associations).First(&file)
	database.Db.Where("file_id = ?", fileID).First(&veriStat)
	database.Db.Where("file_id = ?", fileID).First(&settings)

	go func() {
		fmt.Println("Starting verification")
		fmt.Println("./input/" + file.SystemFileName)
		go core.ValidateCSVFile("./input/"+file.SystemFileName, settings)
		veriStat.Status = "Verifying"
		//ProcessID

		database.Db.Save(&veriStat)
	}()

	//database.Db.First(&file)
	//fmt.Println(file)

}
func UploadFile(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	// single file
	file, err := c.FormFile("file")
	formData, _ := c.MultipartForm()

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
		return
	}
	userID, _ := c.Get("userID")

	var user database.User
	//fmt.Println(userID)
	user.ID = userID.(uint)
	//create uuid for file
	fileName := uuid.New().String()
	var userFile database.File

	userFile.FileName = file.Filename
	userFile.SystemFileName = strconv.Itoa(int(user.ID)) + "-" + fileName + ".csv"
	userFile.UserID = user.ID
	getMissingEmails, _ := strconv.ParseBool(formData.Value["get_missing_emails"][0])
	bruteforceFailedEmails, _ := strconv.ParseBool(formData.Value["brute_force_failed_emails"][0])
	includeGenerics, _ := strconv.ParseBool(formData.Value["include_generic_emails"][0])
	database.Db.Create(&userFile)

	var csvSettings database.CsvFileSettings
	csvSettings.FileID = userFile.ID
	csvSettings.FirstNameColumnName = formData.Value["first_name_column_name"][0]
	csvSettings.LastNameColumnName = formData.Value["last_name_column_name"][0]
	csvSettings.EmailColumnName = formData.Value["email_column_name"][0]
	csvSettings.WebsiteColumnName = formData.Value["domain_column_name"][0]
	csvSettings.GetMissingEmails = getMissingEmails
	csvSettings.BruteForceFailedEmails = bruteforceFailedEmails
	csvSettings.IncludeGenerics = includeGenerics

	database.Db.Create(&csvSettings)

	userFile.Settings.ID = csvSettings.ID

	var verificationStatus database.VerificationStatus
	verificationStatus.FileID = userFile.ID
	verificationStatus.Status = "Not Started"
	database.Db.Create(&verificationStatus)

	userFile.VerificationStatus.ID = verificationStatus.ID

	// Save file to server
	err = c.SaveUploadedFile(file, "./input/"+userFile.SystemFileName)

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
		return
	}

	//userFile.Rows = uint(core.CountRows("./input/" + userFile.SystemFileName))
	userFile.Rows, userFile.HeaderString = core.GetRowsAndHeaders("./input/" + userFile.SystemFileName)
	database.Db.Save(&userFile)
	database.Db.Preload(clause.Associations).First(&user)

	//append assosation for user files
	database.Db.Model(&user).Association("Files").Append(&userFile)
	database.Db.Save(&user)
	//database.Db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user)
	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"files": user.Files,
	})
}
