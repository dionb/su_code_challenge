package storage

import (
	"errors"
	"log"
	"sumup_challenge/api"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type transaction struct {
	ID                 uint64            `gorm:"not null;autoIncrement"`
	SourceAccount      api.AccountNumber `gorm:"not null"`
	DestinationAccount api.AccountNumber `gorm:"not null"`
	Amount             uint64            `gorm:"not null"` //cents
	timestamp          time.Time         `gorm:"not null"`
}

// exported only for easier testing when talking to directly to the db
type Account struct {
	Id      api.AccountNumber `gorm:"primaryKey;not null"`
	Owner   string            `gorm:"not null"`
	Balance uint64            `gorm:"not null"`
}

type PGAccountStore struct{}

func NewPGAccountStore(db *gorm.DB) (*PGAccountStore, error) {
	if err := db.AutoMigrate(&transaction{}, &Account{}); err != nil {
		log.Fatalf("auto-migrate failed: %v", err)
	}
	return new(PGAccountStore), nil
}

func (pas *PGAccountStore) NewAccount(db *gorm.DB, name string, balance uint64) (api.AccountNumber, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	return id, db.Transaction(func(tx *gorm.DB) error {
		if err := db.Create(Account{id, name, balance}).Error; err != nil {
			return err
		}
		trans := newTransaction(uuid.Nil, id, balance)
		return db.Create(&trans).Error
	})
}

func (pas *PGAccountStore) Deposit(db *gorm.DB, accNo api.AccountNumber, amount uint64) (uint64, error) {
	var balance uint64 = 0
	err := db.Transaction(func(tx *gorm.DB) error {
		var acc Account
		if err := tx.Find(&acc, "id = ?", accNo).Error; err != nil {
			return err
		}

		acc.Balance += amount // ignoring overflow. There is not enough money in the world, even allowing for hyperinflation
		balance = acc.Balance

		trans := newTransaction(uuid.Nil, accNo, amount)
		if err := db.Create(&trans).Error; err != nil {
			return err
		}

		return db.Save(&acc).Error
	})

	return balance, err
}

func (pas *PGAccountStore) Transfer(db *gorm.DB, sourceNo, destNo api.AccountNumber, amount uint64) error {
	if sourceNo == destNo {
		return errors.New("Disallowing transfer from an account into itself")
	}
	return db.Transaction(func(tx *gorm.DB) error {
		var source, dest Account

		// query records in a consistent (numerical) order to prevent deadlocks
		if sourceNo.String() < destNo.String() {
			if err := tx.Find(&source, "id = ?", sourceNo).Error; err != nil {
				return err
			}
			if err := tx.Find(&dest, "id = ?", destNo).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Find(&dest, "id = ?", destNo).Error; err != nil {
				return err
			}
			if err := tx.Find(&source, "id = ?", sourceNo).Error; err != nil {
				return err
			}
		}

		// It would be more robust to sum all transactions in and out of the source account to check that
		// it has sufficient funds. But that is overkill given the use of db transactions, and would become
		// expensive for accounts that have a lot of transactions
		if source.Balance < amount {
			return errors.New("Insufficient funds")
		}
		trans := newTransaction(sourceNo, destNo, amount)
		if err := db.Create(&trans).Error; err != nil {
			return err
		}

		source.Balance -= amount
		dest.Balance += amount

		if err := tx.Save(&source).Error; err != nil {
			return err
		}

		if err := tx.Save(&dest).Error; err != nil {
			return err
		}

		return nil
	})
}

func newTransaction(source, destination api.AccountNumber, amount uint64) transaction {
	return transaction{
		0,
		source,
		destination,
		amount,
		time.Now(),
	}
}
