package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccountNumber = uuid.UUID

type AccountStore interface {
	NewAccount(*gorm.DB, string, uint64) (AccountNumber, error)
	Deposit(*gorm.DB, AccountNumber, uint64) (uint64, error)
	Transfer(*gorm.DB, AccountNumber, AccountNumber, uint64) error
}

type ApiServer struct {
	Router       *gin.Engine
	accountStore AccountStore
	db           *gorm.DB
}

func NewServer(db *gorm.DB, store AccountStore) (ApiServer, error) {
	var err error
	server := ApiServer{}
	server.db = db
	server.accountStore = store
	if err != nil {
		return ApiServer{}, err
	}

	server.buildRouter()

	return server, nil
}

func (s *ApiServer) buildRouter() {
	r := gin.Default()
	r.Use(ErrorHandler)
	r.POST("/account", s.handleCreateAccount)
	r.POST("/deposit", s.handleDeposit)
	r.POST("/transfer", s.handleTransfer)
	s.Router = r
}

type CreateAccountRequest struct {
	Name    string
	Balance uint
}

type CreateAccountResponse struct {
	AccountNumber AccountNumber
}

func (s *ApiServer) handleCreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accno, err := s.accountStore.NewAccount(s.db, req.Name, uint64(req.Balance))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, CreateAccountResponse{accno})
}

type DepositRequest struct {
	Account AccountNumber
	Amount  uint32
}

type DepositResponse struct {
	NewBalance uint64
}

func (s *ApiServer) handleDeposit(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	balance, err := s.accountStore.Deposit(s.db, req.Account, uint64(req.Amount))
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, DepositResponse{balance})
}

type TransferRequest struct {
	Source      AccountNumber
	Destination AccountNumber
	Amount      uint32
}

func (s *ApiServer) handleTransfer(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := s.accountStore.Transfer(s.db, req.Source, req.Destination, uint64(req.Amount))
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func ErrorHandler(c *gin.Context) {
	c.Next()

	if len(c.Errors) > 0 {
		err := c.Errors.Last().Err

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal Server Error",
		})

		log.Printf("error: {}", err)
	}
}
