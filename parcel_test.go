package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("DELETE FROM parcel") // Очищаем таблицу перед тестом
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.True(t, id > 0)

	// get
	fetchedParcel, err := store.Get(id)
	require.NoError(t, err)
	parcel.Number = id
	// Устанавливаем CreatedAt из базы для корректного сравнения
	parcel.CreatedAt = fetchedParcel.CreatedAt
	require.Equal(t, parcel, fetchedParcel)

	// delete
	err = store.Delete(id)
	require.NoError(t, err)
	_, err = store.Get(id)
	require.Error(t, err)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("DELETE FROM parcel")
	require.NoError(t, err)
	store := NewParcelStore(db)

	// add
	id, err := store.Add(getTestParcel())
	require.NoError(t, err)
	require.True(t, id > 0)

	// set address
	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	// check
	fetchedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, fetchedParcel.Address)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("DELETE FROM parcel")
	require.NoError(t, err)
	store := NewParcelStore(db)

	// add
	id, err := store.Add(getTestParcel())
	require.NoError(t, err)
	require.True(t, id > 0)

	// set status
	newStatus := "sent"
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	// check
	fetchedParcel, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, fetchedParcel.Status)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db")
	require.NoError(t, err)
	defer db.Close()
	_, err = db.Exec("DELETE FROM parcel")
	require.NoError(t, err)
	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	for i := range parcels {
		parcels[i].Client = client
	}

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		require.True(t, id > 0)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		original, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		// Устанавливаем CreatedAt из базы для корректного сравнения
		original.CreatedAt = parcel.CreatedAt
		require.Equal(t, original, parcel)
	}
}
