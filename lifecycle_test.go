package main

import (
	"sumup_challenge/api"
	"sumup_challenge/api/client"
	"sumup_challenge/storage"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestLifecycle(t *testing.T) {
	var err error
	c, err := client.New("http://localhost:8080")
	assert.NoError(t, err)

	db := connectDB()

	var AccountNoA, AccountNoB api.AccountNumber

	t.Run("create account with 0 balance", func(t *testing.T) {
		AccountNoA, err = c.CreateAccount("A", 0)
		assert.NoError(t, err)

		AccountA, err := GetAccount(db, AccountNoA)
		assert.NoError(t, err)

		assert.Equal(t, "A", AccountA.Owner)
		assert.Equal(t, uint64(0), AccountA.Balance)
	})

	t.Run("Create an account with a balance of 100", func(t *testing.T) {
		AccountNoB, err = c.CreateAccount("B", 100)
		assert.NoError(t, err)

		AccountB, err := GetAccount(db, AccountNoB)
		assert.NoError(t, err)

		assert.Equal(t, "B", AccountB.Owner)
		assert.Equal(t, uint64(100), AccountB.Balance)
	})

	// Can't create an account with a negative balance to check the api rejects it. A shortsight in this test
	// framework via sdk approach. It would be worth creating a separate test for this in a production system

	t.Run("Create a deposit in account A", func(t *testing.T) {
		nb, err := c.CreateDeposit(AccountNoA, 1)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), nb)

		Account, err := GetAccount(db, AccountNoA)
		assert.NoError(t, err)

		assert.Equal(t, "A", Account.Owner)
		assert.Equal(t, uint64(1), Account.Balance)
	})

	// Can't create an deposit with a negative balance to check the api rejects it. A shortsight in this test
	// framework via sdk approach. It would be worth creating a separate test for this in a production system

	t.Run("Unsuccessful transfer", func(t *testing.T) {
		beforeA, err := GetAccount(db, AccountNoA)
		assert.NoError(t, err)
		beforeB, err := GetAccount(db, AccountNoB)
		assert.NoError(t, err)

		err = c.RequestTransfer(AccountNoA, AccountNoB, 2)
		assert.Error(t, err)

		afterA, err := GetAccount(db, AccountNoA)
		assert.NoError(t, err)
		afterB, err := GetAccount(db, AccountNoB)
		assert.NoError(t, err)
		assert.Equal(t, beforeA.Owner, afterA.Owner)
		assert.Equal(t, beforeA.Balance, afterA.Balance)
		assert.Equal(t, beforeB.Owner, afterB.Owner)
		assert.Equal(t, beforeB.Balance, afterB.Balance)
	})

	t.Run("Successful transfer", func(t *testing.T) {
		const amount = 3

		beforeA, err := GetAccount(db, AccountNoA)
		assert.NoError(t, err)
		beforeB, err := GetAccount(db, AccountNoB)
		assert.NoError(t, err)

		err = c.RequestTransfer(AccountNoB, AccountNoA, 3)
		assert.NoError(t, err)

		afterA, err := GetAccount(db, AccountNoA)
		assert.NoError(t, err)
		afterB, err := GetAccount(db, AccountNoB)
		assert.NoError(t, err)
		assert.Equal(t, beforeA.Owner, afterA.Owner)
		assert.Equal(t, beforeA.Balance+amount, afterA.Balance)
		assert.Equal(t, beforeB.Owner, afterB.Owner)
		assert.Equal(t, beforeB.Balance-amount, afterB.Balance)

		// TODO check that the transaction record was created correctly
	})

}
func GetAccount(db *gorm.DB, AccNo api.AccountNumber) (storage.Account, error) {
	var acc storage.Account
	err := db.Find(&acc, "id = ?", AccNo).Error
	return acc, err
}
