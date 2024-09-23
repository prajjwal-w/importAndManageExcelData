package controller

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/prajjwal-w/golang-choicetech/helpers"
	"github.com/prajjwal-w/golang-choicetech/model"
	"github.com/prajjwal-w/golang-choicetech/service"
)

func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {

		file, err := c.FormFile("file")

		if err != nil {
			log.Println("Invalid File")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid File"})
			return
		}

		//save the uploaded file
		filePath := filepath.Join("uploads", file.Filename)
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			log.Println("error while saving the file")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while saving the file"})
			return
		}

		//validate the uploaded file in terms of our struct person
		var person model.Person
		excel_data, err := helpers.ValidateExcel(filePath, person)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while validating excel"})
			return
		}

		//parseing the data and storing it into the db async
		persons := helpers.ParsingData(excel_data)
		if persons == nil {
			log.Println("error while parsing")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while parsing the data"})
			return
		}

		// Start a goroutine to process the data
		go func(persons []model.Person) {
			if err := service.StoreData(persons); err != nil {
				log.Printf("Error storing data: %v", err)
			}
		}(persons)

		go func(persons []model.Person) {
			if err := service.CacheData(persons); err != nil {
				log.Printf("Error caching data:  %v", err)
			}
		}(persons)

		log.Println("len of the perons array: ", len(persons))
		c.JSON(http.StatusOK, gin.H{"message": "You data is being processed"})

	}
}

func ViewData() gin.HandlerFunc {
	return func(c *gin.Context) {
		email := c.Param("email")

		if len(email) > 0 && email[0] == '/' {
			email = email[1:]
		}

		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		//get the data from redis or databse
		data, err := service.GetTheDataByEmail(email)
		if err != nil {
			log.Printf("error while querying the data : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"person": data})

	}
}

func UpdateData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var person *model.Person

		if err := c.BindJSON(&person); err != nil {
			log.Println("error while binding JSON")
			c.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}
		log.Println(person)

		//update data into the databse and redis
		err := service.UpdateData(person)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Record Updated Successfully"})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "You have logged-in"})
	}
}
