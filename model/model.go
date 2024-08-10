package model

//Person Model

type Person struct {
	FirstName   string `json:"first_name" gorm:"column:first_name"`
	LastName    string `json:"last_name" gorm:"column:last_name"`
	CompanyName string `json:"company_name" gorm:"column:company_name"`
	Address     string `json:"address" gorm:"column:address"`
	City        string `json:"city" gorm:"column:city"`
	County      string `json:"county" gorm:"column:county"`
	Postal      string `json:"postal" gorm:"column:postal"`
	Phone       string `json:"phone" gorm:"column:phone"`
	Email       string `json:"email" gorm:"column:email"`
	Web         string `json:"web" gorm:"column:web"`
}
