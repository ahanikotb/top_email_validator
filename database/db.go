package database

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

type User struct {
	gorm.Model
	FirstName string
	LastName  string
	Email     string
	Password  string
	Files     []File
}

type File struct {
	gorm.Model
	UserID             uint
	FileName           string
	SystemFileName     string
	VerificationStatus VerificationStatus `gorm:"embedded"`
	ProcessID          uint
	Rows               uint
	Settings           CsvFileSettings `gorm:"embedded"`
	HeaderString       string          //header string for csv file
}
type CsvFileSettings struct {
	gorm.Model
	FileID                 uint
	FirstNameColumnName    string `json:"first_name_column_name"`
	LastNameColumnName     string `json:"last_name_column_name"`
	EmailColumnName        string `json:"email_column_name"`
	WebsiteColumnName      string `json:"domain_column_name"`
	GetMissingEmails       bool   `json:"get_missing_emails"`
	BruteForceFailedEmails bool   `json:"brute_force_failed_emails"`
	IncludeGenerics        bool   `json:"include_generics"`
}
type VerificationStatus struct {
	gorm.Model
	FileID          uint
	Status          string
	PercentageDone  uint
	ValidEmails     uint
	CatchAllEmails  uint
	FullInboxEmails uint
	InvalidEmails   uint
}

var Db *gorm.DB

//	func OPENDB() *gorm.DB {
//		db, _ := gorm.Open(sqlite.Open("./database/test.db"), &gorm.Config{Logger: logger.Default})
//		err := db.AutoMigrate(&User{}, &File{}, &VerificationStatus{})
//		if err != nil {
//			fmt.Println(err.Error())
//		}
//		db.Set("gorm:auto_preload", true)
//		return db
//	}
func OPENDB() {

	_db, _ := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{Logger: logger.Default})
	_db.AutoMigrate(&User{}, &File{}, &VerificationStatus{}, &CsvFileSettings{})

	//Delete all records
	_db.Delete(&User{}, "id > 0")
	_db.Delete(&File{}, "id > 0")
	_db.Delete(&VerificationStatus{}, "id > 0")
	_db.Delete(&CsvFileSettings{}, "id > 0")

	var user User

	user.FirstName = "user"
	user.LastName = "user"

	user.Email = "email"

	hpass, _ := bcrypt.GenerateFromPassword([]byte("pass"), 10)

	user.Password = string(hpass)

	_db.Save(&user)
	//_db.Set("gorm:auto_preload", true)
	Db = _db
}
