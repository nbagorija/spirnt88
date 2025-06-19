package main

import (
	"database/sql"
	"errors"
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
		if errors.Is(err, sql.ErrNoRows) {
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

	// Проверяем ошибки, возникшие при переборе строк
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	query := "UPDATE parcel SET status = ? WHERE number = ?"
	_, err := s.db.Exec(query, status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	query := "UPDATE parcel SET address = ? WHERE number = ? AND status = 'registered'"
	result, err := s.db.Exec(query, address, number)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("address can only be updated for registered parcels")
	}
	return nil
}

func (s ParcelStore) Delete(number int) error {
	query := "DELETE FROM parcel WHERE number = ? AND status = 'registered'"
	result, err := s.db.Exec(query, number)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("can only delete registered parcels")
	}
	return nil
}
