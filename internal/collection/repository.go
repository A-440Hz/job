package collection

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const seedFile = "data/collectables.json"
const pathToCollectables = "pathToCollectables"

func init() {
	os.Setenv(pathToCollectables, "")
}

func SetEnvForTesting(dir string) {
	os.Setenv(pathToCollectables, filepath.ToSlash(filepath.Join(dir, seedFile)))
}

func getSeedFilePath() string {

	if p := os.Getenv(pathToCollectables); p != "" {
		return p
	}

	// this has been stinky bad code but it was fine because it worked.
	// this does not work with a built binary. It will loop forever unless it is placed in a specific location.
	// TODO: make this better
	baseDir, err := os.Executable()
	if err != nil {
		return seedFile
	}
	for filepath.Base(baseDir) != "internal" && filepath.Base(baseDir) != "cmd" {
		baseDir = filepath.Clean(filepath.Join(baseDir, "../"))
	}
	return filepath.ToSlash(filepath.Join(filepath.Dir(baseDir), seedFile))
}

type Repository struct {
	db   *gorm.DB
	size int
}

func NewRepository(d *gorm.DB) *Repository {
	return &Repository{db: d}
}

// importCollectables updates the collectables table by re-importing it from the json seedFile
func (r *Repository) importCollectables() error {
	f, err := os.Open(getSeedFilePath())
	if err != nil {
		return err
	}
	defer f.Close()
	var collectables []Collectable
	err = json.NewDecoder(f).Decode(&collectables)
	if err != nil {
		return err
	}
	r.size = len(collectables)
	for _, c := range collectables {
		res := r.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(&c)
		if res.Error != nil {
			return res.Error
		}
	}
	return nil
}

func (u *UserInventory) BeforeCreate(tx *gorm.DB) error {
	u.LastCompleted = time.Time{} // set to zero value
	return nil
}

func (r *Repository) createUserInventory(u *UserInventory) (*UserInventory, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) lookupUserInventory(id string) (*UserInventory, error) {
	u := &UserInventory{UserID: id}
	res := r.db.First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) updateUserInventoryFields(u *UserInventory, fields []string) (*UserInventory, error) {
	res := r.db.Model(u).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) deleteUserInventory(u *UserInventory) error {
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user inventory not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) beginCollectablesTransaction() *gorm.DB {
	return r.db.Begin()
}

func (r *Repository) createUserCollectable(u *UserCollectable) (*UserCollectable, error) {
	res := r.db.Create(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) updateUserCollectableFields(u *UserCollectable, fields []string) (*UserCollectable, error) {
	res := r.db.Model(u).Where("user_id = ? AND collectable_id = ?", u.UserID, u.CollectableID).Select(fields).Updates(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) lookupUserCollectable(userId string, collectableID int) (*UserCollectable, error) {
	u := &UserCollectable{}
	res := r.db.Where("user_id = ?", userId).
		Where("collectable_id = ?", collectableID).
		Preload("Collectable").
		First(u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) getAllCollectablesForUser(userId string) ([]UserCollectable, error) {
	u := []UserCollectable{}
	res := r.db.Where("user_id = ?", userId).Preload("Collectable", func(db *gorm.DB) *gorm.DB {
		return db.Order("collectables.ID ASC")
	}).Find(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	return u, nil
}

func (r *Repository) deleteAllUserCollectables(userId string) error {
	u := []UserCollectable{}
	res := r.db.Where("user_id = ?", userId).Preload("Collectable").Delete(&u)
	return res.Error
}

func (r *Repository) deleteUserCollectable(u *UserCollectable) error {
	res := r.db.Delete(u)
	if res.RowsAffected == 0 {
		return errors.New("user collectable not found")
	}
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (r *Repository) selectAllCollectables() ([]Collectable, error) {
	collectables := []Collectable{}
	res := r.db.Find(&collectables)
	if res.Error != nil {
		return nil, res.Error
	}
	return collectables, nil
}
