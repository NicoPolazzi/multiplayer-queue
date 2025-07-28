package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code   string
	Price  uint
	UserID uint
}

type User struct {
	gorm.Model
	Name    string
	Product Product
}

type SuiteTester struct {
	suite.Suite
	db *gorm.DB
}

// This runs BEFORE all tests
func (s *SuiteTester) SetupSuite() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	s.db = db
}

// This runs AFTER all tests
func (s *SuiteTester) TearDownSuite() {
	db, _ := s.db.DB()
	db.Close()
}

func (s *SuiteTester) SetupTest() {
	s.db.Migrator().DropTable(&Product{})
	s.db.AutoMigrate(&Product{}, &User{})
}

func (s *SuiteTester) TestCreate() {
	product := Product{Code: "D42", Price: 100}

	s.db.Create(&product)
	assert.Equal(s.T(), uint(1), product.ID)
}

func (s *SuiteTester) TestBatchCreate() {
	products := []*Product{{Code: "D42", Price: 100}, {Code: "D43", Price: 200}}

	s.db.Create(products)

	assert.Equal(s.T(), uint(1), products[0].ID)
	assert.Equal(s.T(), uint(2), products[1].ID)
}

func (s *SuiteTester) TestCreateWithAssociations() {
	user := User{Name: "test", Product: Product{Code: "D42", Price: 100}}

	s.db.Create(&user)
	assert.Equal(s.T(), uint(1), user.Product.ID)
	assert.Equal(s.T(), uint(1), user.ID)
}

func (s *SuiteTester) TestFirst() {
	products := []*Product{{Code: "D42", Price: 100}, {Code: "D43", Price: 200}}
	p := Product{}

	s.db.Create(products)
	s.db.First(&p)

	assert.Equal(s.T(), "D42", p.Code)
	assert.Equal(s.T(), uint(100), p.Price)
}

func (s *SuiteTester) TestLast() {
	products := []*Product{{Code: "D42", Price: 100}, {Code: "D43", Price: 200}}
	p := Product{}

	s.db.Create(products)
	s.db.Last(&p)

	assert.Equal(s.T(), "D43", p.Code)
	assert.Equal(s.T(), uint(200), p.Price)
}

func (s *SuiteTester) TestFind() {
	product1 := Product{Code: "D42", Price: 100}
	product2 := Product{Code: "D43", Price: 200}
	var retrievedProducts []Product

	s.db.Create(&product1)
	s.db.Create(&product2)

	result := s.db.Find(&retrievedProducts)

	assert.Equal(s.T(), int64(2), result.RowsAffected)
}

func (s *SuiteTester) TestConditions() {
	products := []*Product{{Code: "D42", Price: 100}, {Code: "D43", Price: 200}}
	s.db.Create(products)

	var retrievedProduct Product

	result := s.db.Where(&Product{Code: "D42", Price: 300}).First(&retrievedProduct)
	assert.NotNil(s.T(), result.Error)

	result = s.db.Where(&Product{Code: "D43", Price: 200}).First(&retrievedProduct)
	assert.Nil(s.T(), result.Error)
	assert.True(s.T(), retrievedProduct.Code == "D43" && retrievedProduct.Price == 200)
}

func (s *SuiteTester) TestSave() {
	product1 := Product{Code: "D42", Price: 100}
	s.db.Save(&product1)

	result := s.db.First(&product1)

	assert.True(s.T(), result.RowsAffected == 1)
	assert.Nil(s.T(), result.Error)

	s.db.Save(&Product{Model: gorm.Model{ID: 1}, Code: "D42", Price: 200})
	s.db.First(&product1)

	assert.Equal(s.T(), uint(200), product1.Price)
}

func (s *SuiteTester) TestUpdate() {
	product := Product{Code: "D43", Price: 200}
	s.db.Create(&product)

	s.db.Model(&product).Update("code", "D42")
	s.db.Model(&product).Update("price", "400")

	assert.Equal(s.T(), "D42", product.Code)
	assert.Equal(s.T(), uint(400), product.Price)

}

func TestSuiteTester(t *testing.T) {
	suite.Run(t, new(SuiteTester))
}
