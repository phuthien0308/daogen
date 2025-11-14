package dto

type Person struct {
	ID        int    `sql-col:"id" sql-identity:"true"`
	FirstName string `sql-col:"first_name"`
	LastName  string `sql-col:"last_name"`
	Email     string `sql-col:"email" sql-update:"false"`
}
