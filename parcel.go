package main

import (
	"database/sql"
	"fmt"
	"time"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	query := "INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)"
	result, err := s.db.Exec(query, p.Client, "registered", p.Address, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE number = ?"
	var p Parcel
	err := s.db.QueryRow(query, number).Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, fmt.Errorf("parcel not found")
		}
		return p, err
	}
	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	query := "SELECT number, client, status, address, created_at FROM parcel WHERE client = ?"
	rows, err := s.db.Query(query, client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel
	for rows.Next() {
		var p Parcel
		err = rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, rows.Err()
}

func (s ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel SET status = ? WHERE number = ?"
	_, err := s.db.Exec(query, status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	// Проверяем статус перед обновлением
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel not found")
		}
		return err
	}
	if currentStatus != "registered" {
		return fmt.Errorf("address can only be updated for registered parcels")
	}

	query := "UPDATE parcel SET address = ? WHERE number = ?"
	_, err = s.db.Exec(query, address, number)
	return err
}

func (s ParcelStore) Delete(number int) error {
	// Проверяем статус перед удалением
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("parcel not found")
		}
		return err
	}
	if currentStatus != "registered" {
		return fmt.Errorf("can only delete registered parcels")
	}

	query := "DELETE FROM parcel WHERE number = ?"
	_, err = s.db.Exec(query, number)
	return err
}
