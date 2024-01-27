package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	// настройте подключение к БД

	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	parcelId, err := store.Add(parcel)

	require.NoError(t, err, "test add no error")
	require.NotEmpty(t, parcelId, "Test id is present")

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	testParcel, err := store.Get(parcelId)
	require.NoError(t, err, "test get no error")
	require.Equal(t, parcel.Client, testParcel.Client, "Test parcel.Client")
	require.Equal(t, parcel.Status, testParcel.Status, "Test parcel.Status")
	require.Equal(t, parcel.Address, testParcel.Address, "Test parcel.Address")
	require.Equal(t, parcel.CreatedAt, testParcel.CreatedAt, "Test parcel.CreatedAt")

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(parcelId)
	require.NoError(t, err, "test delete no error")

	_, err = store.Get(parcelId)
	require.Error(t, sql.ErrNoRows, err, "test get with error")
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	parcelId, err := store.Add(parcel)

	require.NoError(t, err, "test add no error")
	require.NotEmpty(t, parcelId, "Test id is present")

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(parcelId, newAddress)

	require.NoError(t, err, "test set address no error")

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	testParcel, err := store.Get(parcelId)
	require.NoError(t, err, "test get no error")
	require.Equal(t, newAddress, testParcel.Address, "test address is updated")
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	parcelId, err := store.Add(parcel)

	require.NoError(t, err, "test add no error")
	require.NotEmpty(t, parcelId, "Test id is present")

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(parcelId, ParcelStatusSent)
	require.NoError(t, err, "Test set status no error")

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	testParcel, err := store.Get(parcelId)
	require.NoError(t, err, "test get no error")
	require.Equal(t, ParcelStatusSent, testParcel.Status, "test status is updated")
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	// настройте подключение к БД
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := range parcels {
		// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		id, err := store.Add(parcels[i])
		require.NoError(t, err, "Test id is present")
		require.NotEmpty(t, id, "Test id is present")
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	// получите список посылок по идентификатору клиента, сохранённого в переменной client
	storedParcels, err := store.GetByClient(client)
	// убедитесь в отсутствии ошибки
	require.NoError(t, err, "Test get no error")
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Equal(t, len(parcels), len(storedParcels))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		p, ok := parcelMap[parcel.Number]
		require.True(t, ok, "test found by id")
		require.Equal(t, p, parcel, "test parcel is equal")
	}
}
