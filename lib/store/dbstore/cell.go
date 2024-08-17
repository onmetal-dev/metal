package dbstore

import (
	"github.com/onmetal-dev/metal/lib/store"
	"go.jetify.com/typeid"
	"gorm.io/gorm"
)

type CellStore struct {
	db *gorm.DB
}

var _ store.CellStore = &CellStore{}

type NewCellStoreParams struct {
	DB *gorm.DB
}

func NewCellStore(params NewCellStoreParams) *CellStore {
	return &CellStore{
		db: params.DB,
	}
}

func (s *CellStore) Create(cell store.Cell) (store.Cell, error) {
	tid, _ := typeid.WithPrefix("cell")
	cell.Id = tid.String()
	if cell.TalosCellData != nil {
		cell.TalosCellData.CellId = cell.Id
	}
	return cell, s.db.Create(&cell).Error
}

func (s *CellStore) Get(id string) (store.Cell, error) {
	var cell store.Cell
	return cell, s.db.Preload("Servers").Preload("TalosCellData").Where("id = ?", id).First(&cell).Error
}

func (s *CellStore) GetForTeam(teamId string) ([]store.Cell, error) {
	var cells []store.Cell
	return cells, s.db.Preload("Servers").Preload("TalosCellData").Where("team_id = ?", teamId).Find(&cells).Error
}

func (s *CellStore) UpdateTalosCellData(cellId string, talosCellData store.TalosCellData) error {
	return s.db.Model(&store.Cell{}).Where("id = ?", cellId).Association("TalosCellData").Replace(&talosCellData)
}

func (s *CellStore) AddServer(cellId string, server store.Server) error {
	return s.db.Model(&store.Cell{}).Where("id = ?", cellId).Association("Servers").Append(&server)
}
