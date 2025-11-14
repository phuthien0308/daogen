package storage

import (
	"context"
	"database/sql"
	"fmt"

	// external imports
	dto "github.com/phuthien0308/daogen/dto"
)

// -----------------------------------------------------------------
// 1. DAO Interface (The Contract)
// -----------------------------------------------------------------

// PersonDAO defines the methods for interacting with Person data in the database.
type PersonDAO interface {
	Create(ctx context.Context, person *dto.Person) (int64, error)
}

// -----------------------------------------------------------------
// 2. Concrete Implementation Struct
// -----------------------------------------------------------------

// personDAOImpl holds the database connection and implements the PersonDAO interface.
type personDAOImpl struct {
	db *sql.DB
}

// NewPersonDAO creates and returns a new PersonDAO implementation.
func NewPersonDAO(db *sql.DB) PersonDAO {
	return &personDAOImpl{db: db}
}

// Create inserts a new person record into the database.
func (dao *personDAOImpl) Create(ctx context.Context, person *dto.Person) (int64, error) {
	// WARNING: You must adjust the number of $N placeholders and fields in .Scan()
	query := "INSERT INTO people (first_name,last_name,email) VALUES (?,?,?)"

	// Example: .Scan(&person.ID, &person.Field1, &person.Field2)

	result, err := dao.db.ExecContext(ctx, query, person.FirstName, person.LastName, person.Email)

	if err != nil {
		return 0, fmt.Errorf("error creating person: %w", err)
	}
	return result.LastInsertId()
}
