package main

import (
	"topemailvalidator/api"
	"topemailvalidator/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Println("Recovered in f", r)
		}
	}()
	godotenv.Load()
	database.OPENDB()
	//if err := os.Remove("./database/test.db"); err != nil {
	//	log.Fatal(err)
	//}
	//src := "./database/backup.db"
	//dest := "./database/test.db"
	//
	//bytesRead, err := ioutil.ReadFile(src)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//err = ioutil.WriteFile(dest, bytesRead, 0644)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}

	r := gin.Default()
	r.Use(api.CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	api.MakeRoutes(r)
	r.Run()
	// wait for all the goroutines to finish
}

//func splitCsv(filename string, division int) {
//	csvFile := core.LoadCsv(filename)
//
//	div := len(csvFile) / division
//
//	div = int(math.Round(float64(div)))
//	//loop through the csvFile and split it into 10 files
//	for i := 0; i < division; i++ {
//		number := strconv.Itoa(i + 1)
//		filename := strings.Join([]string{"./output/engineeringlist", number, ".csv"}, "")
//		core.SaveCsv(filename, csvFile[i*div:(i+1)*div])
//
//	}
//}
