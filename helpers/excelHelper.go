package helpers

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/prajjwal-w/golang-choicetech/model"
	"github.com/xuri/excelize/v2"
)

func ValidateExcel(filepath string, structs interface{}) ([][]string, error) {
	excelFile, err := excelize.OpenFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error while opening the excelfile: %v", err)
	}

	defer excelFile.Close()

	//get the first sheet name
	sheet := excelFile.GetSheetName(0)

	rows, err := excelFile.GetRows(sheet)
	if err != nil || len(rows) == 0 {
		log.Printf("len of rows :%v", len(rows))
		return nil, fmt.Errorf("error while getting the rows or the file is empty")
	}

	//validating the header row with the Person struct to check it is same or not
	tags := make(map[string]string)
	val := reflect.ValueOf(structs)
	typ := reflect.TypeOf(structs)

	//store the column name into a map
	for i := 0; i < val.NumField(); i++ {
		f := typ.Field(i)
		c_tag := f.Tag.Get("gorm")
		if strings.HasPrefix(c_tag, "column:") {
			col_name := strings.TrimPrefix(c_tag, "column:")
			tags[strings.ToLower(col_name)] = f.Name
		}
	}

	//store the header row form the excel file into a slice
	rowHeader := make([]string, 0)
	for _, v := range rows[0] {
		rowHeader = append(rowHeader, strings.ToLower(v))
	}

	//checking the column and the struct fields
	for col_Name := range tags {
		found := false
		for _, header := range rowHeader {
			if col_Name == header {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("columns header %s is not found in the struct", col_Name)
		}
	}
	return rows, nil
}

func ParsingData(data [][]string) []model.Person {
	t := time.Now()
	var person []model.Person
	var wg sync.WaitGroup

	rows := make(chan []string)
	result := make(chan model.Person)

	//number of the worker routine
	const worker = 20

	//start the worker routine
	for i := 0; i < worker; i++ {
		wg.Add(1)
		go parseRows(rows, result, &wg)
	}

	go func() {
		for _, r := range data[1:] {
			rows <- r
		}
		close(rows)
	}()

	go func() {
		wg.Wait()
		close(result)
	}()

	for res := range result {
		person = append(person, res)
	}
	log.Printf("Parsring time: %v", time.Since(t))
	return person

}

func parseRows(rows <-chan []string, result chan<- model.Person, wg *sync.WaitGroup) {
	defer wg.Done()
	for r := range rows {
		p := model.Person{}
		p.FirstName = r[0]
		p.LastName = r[1]
		p.CompanyName = r[2]
		p.Address = r[3]
		p.City = r[4]
		p.County = r[5]
		p.Postal = r[6]
		p.Phone = r[7]
		p.Email = r[8]
		p.Web = r[9]

		result <- p
	}
}
