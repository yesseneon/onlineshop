package brand

import (
	"database/sql"
	"onlineshop/config"
)

// Brand struct
type Brand struct {
	ID        int    `db:"id"`
	Name      string `db:"name"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (b *Brand) store() (int, error) {
	var lastInsertedID int
	sqlStatement := "INSERT INTO brands (name, created_at, updated_at) VALUES ($1, NOW()::timestamp(0), NOW()::timestamp(0)) RETURNING id"
	err := config.DB.QueryRow(sqlStatement, b.Name).Scan(&lastInsertedID)
	if err != nil {
		return lastInsertedID, err
	}
	return lastInsertedID, nil
}

func (b *Brand) update() error {
	_, err := config.DB.Exec("UPDATE brands SET name=$1, updated_at=NOW()::timestamp(0) WHERE id=$2", b.Name, b.ID)
	if err != nil {
		return err
	}
	return nil
}

func (b *Brand) destroy() error {
	_, err := config.DB.Exec("DELETE FROM brands WHERE id=$1", b.ID)
	if err != nil {
		return err
	}
	return nil
}

func allBrands() ([]Brand, error) {
	rows, err := config.DB.Query("SELECT * FROM brands")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	brands := []Brand{}
	for rows.Next() {
		brand := Brand{}
		err := rows.Scan(&brand.ID, &brand.Name, &brand.CreatedAt, &brand.UpdatedAt)
		if err != nil {
			return nil, err
		}
		brands = append(brands, brand)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return brands, nil
}

func oneBrand(id int) (Brand, error) {
	brand := Brand{}
	err := config.DB.QueryRow("SELECT * FROM brands WHERE id=$1", id).Scan(&brand.ID, &brand.Name, &brand.CreatedAt, &brand.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return brand, nil
		}
		return brand, err
	}
	return brand, nil
}