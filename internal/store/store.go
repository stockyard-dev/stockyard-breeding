package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Animals struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Species string `json:"species"`
	Breed string `json:"breed"`
	Sex string `json:"sex"`
	DateOfBirth string `json:"date_of_birth"`
	RegistrationNumber string `json:"registration_number"`
	SireId string `json:"sire_id"`
	DamId string `json:"dam_id"`
	Color string `json:"color"`
	Status string `json:"status"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

type Litters struct {
	ID string `json:"id"`
	SireId string `json:"sire_id"`
	DamId string `json:"dam_id"`
	DateBorn string `json:"date_born"`
	Count int64 `json:"count"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

type HealthRecords struct {
	ID string `json:"id"`
	AnimalId string `json:"animal_id"`
	RecordType string `json:"record_type"`
	Date string `json:"date"`
	Description string `json:"description"`
	VetName string `json:"vet_name"`
	Cost float64 `json:"cost"`
	Notes string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil { return nil, err }
	db, err := sql.Open("sqlite", filepath.Join(d, "breeding.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil { return nil, err }
	db.SetMaxOpenConns(1)
	db.Exec(`CREATE TABLE IF NOT EXISTS animals(id TEXT PRIMARY KEY, name TEXT NOT NULL, species TEXT DEFAULT '', breed TEXT DEFAULT '', sex TEXT DEFAULT '', date_of_birth TEXT DEFAULT '', registration_number TEXT DEFAULT '', sire_id TEXT DEFAULT '', dam_id TEXT DEFAULT '', color TEXT DEFAULT '', status TEXT DEFAULT '', notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS litters(id TEXT PRIMARY KEY, sire_id TEXT NOT NULL, dam_id TEXT NOT NULL, date_born TEXT NOT NULL, count INTEGER DEFAULT 0, notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS health_records(id TEXT PRIMARY KEY, animal_id TEXT NOT NULL, record_type TEXT DEFAULT '', date TEXT NOT NULL, description TEXT DEFAULT '', vet_name TEXT DEFAULT '', cost REAL DEFAULT 0, notes TEXT DEFAULT '', created_at TEXT DEFAULT(datetime('now')))`)
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (d *DB) CreateAnimals(e *Animals) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO animals(id, name, species, breed, sex, date_of_birth, registration_number, sire_id, dam_id, color, status, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.Name, e.Species, e.Breed, e.Sex, e.DateOfBirth, e.RegistrationNumber, e.SireId, e.DamId, e.Color, e.Status, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetAnimals(id string) *Animals {
	var e Animals
	if d.db.QueryRow(`SELECT id, name, species, breed, sex, date_of_birth, registration_number, sire_id, dam_id, color, status, notes, created_at FROM animals WHERE id=?`, id).Scan(&e.ID, &e.Name, &e.Species, &e.Breed, &e.Sex, &e.DateOfBirth, &e.RegistrationNumber, &e.SireId, &e.DamId, &e.Color, &e.Status, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListAnimals() []Animals {
	rows, _ := d.db.Query(`SELECT id, name, species, breed, sex, date_of_birth, registration_number, sire_id, dam_id, color, status, notes, created_at FROM animals ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Animals
	for rows.Next() { var e Animals; rows.Scan(&e.ID, &e.Name, &e.Species, &e.Breed, &e.Sex, &e.DateOfBirth, &e.RegistrationNumber, &e.SireId, &e.DamId, &e.Color, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateAnimals(e *Animals) error {
	_, err := d.db.Exec(`UPDATE animals SET name=?, species=?, breed=?, sex=?, date_of_birth=?, registration_number=?, sire_id=?, dam_id=?, color=?, status=?, notes=? WHERE id=?`, e.Name, e.Species, e.Breed, e.Sex, e.DateOfBirth, e.RegistrationNumber, e.SireId, e.DamId, e.Color, e.Status, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteAnimals(id string) error {
	_, err := d.db.Exec(`DELETE FROM animals WHERE id=?`, id)
	return err
}

func (d *DB) CountAnimals() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM animals`).Scan(&n); return n
}

func (d *DB) SearchAnimals(q string, filters map[string]string) []Animals {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR breed LIKE ? OR registration_number LIKE ? OR sire_id LIKE ? OR dam_id LIKE ? OR color LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["species"]; ok && v != "" { where += " AND species=?"; args = append(args, v) }
	if v, ok := filters["sex"]; ok && v != "" { where += " AND sex=?"; args = append(args, v) }
	if v, ok := filters["status"]; ok && v != "" { where += " AND status=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, name, species, breed, sex, date_of_birth, registration_number, sire_id, dam_id, color, status, notes, created_at FROM animals WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Animals
	for rows.Next() { var e Animals; rows.Scan(&e.ID, &e.Name, &e.Species, &e.Breed, &e.Sex, &e.DateOfBirth, &e.RegistrationNumber, &e.SireId, &e.DamId, &e.Color, &e.Status, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateLitters(e *Litters) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO litters(id, sire_id, dam_id, date_born, count, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?)`, e.ID, e.SireId, e.DamId, e.DateBorn, e.Count, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetLitters(id string) *Litters {
	var e Litters
	if d.db.QueryRow(`SELECT id, sire_id, dam_id, date_born, count, notes, created_at FROM litters WHERE id=?`, id).Scan(&e.ID, &e.SireId, &e.DamId, &e.DateBorn, &e.Count, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListLitters() []Litters {
	rows, _ := d.db.Query(`SELECT id, sire_id, dam_id, date_born, count, notes, created_at FROM litters ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []Litters
	for rows.Next() { var e Litters; rows.Scan(&e.ID, &e.SireId, &e.DamId, &e.DateBorn, &e.Count, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateLitters(e *Litters) error {
	_, err := d.db.Exec(`UPDATE litters SET sire_id=?, dam_id=?, date_born=?, count=?, notes=? WHERE id=?`, e.SireId, e.DamId, e.DateBorn, e.Count, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteLitters(id string) error {
	_, err := d.db.Exec(`DELETE FROM litters WHERE id=?`, id)
	return err
}

func (d *DB) CountLitters() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM litters`).Scan(&n); return n
}

func (d *DB) SearchLitters(q string, filters map[string]string) []Litters {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (sire_id LIKE ? OR dam_id LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	rows, _ := d.db.Query(`SELECT id, sire_id, dam_id, date_born, count, notes, created_at FROM litters WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []Litters
	for rows.Next() { var e Litters; rows.Scan(&e.ID, &e.SireId, &e.DamId, &e.DateBorn, &e.Count, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) CreateHealthRecords(e *HealthRecords) error {
	e.ID = genID(); e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO health_records(id, animal_id, record_type, date, description, vet_name, cost, notes, created_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`, e.ID, e.AnimalId, e.RecordType, e.Date, e.Description, e.VetName, e.Cost, e.Notes, e.CreatedAt)
	return err
}

func (d *DB) GetHealthRecords(id string) *HealthRecords {
	var e HealthRecords
	if d.db.QueryRow(`SELECT id, animal_id, record_type, date, description, vet_name, cost, notes, created_at FROM health_records WHERE id=?`, id).Scan(&e.ID, &e.AnimalId, &e.RecordType, &e.Date, &e.Description, &e.VetName, &e.Cost, &e.Notes, &e.CreatedAt) != nil { return nil }
	return &e
}

func (d *DB) ListHealthRecords() []HealthRecords {
	rows, _ := d.db.Query(`SELECT id, animal_id, record_type, date, description, vet_name, cost, notes, created_at FROM health_records ORDER BY created_at DESC`)
	if rows == nil { return nil }; defer rows.Close()
	var o []HealthRecords
	for rows.Next() { var e HealthRecords; rows.Scan(&e.ID, &e.AnimalId, &e.RecordType, &e.Date, &e.Description, &e.VetName, &e.Cost, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}

func (d *DB) UpdateHealthRecords(e *HealthRecords) error {
	_, err := d.db.Exec(`UPDATE health_records SET animal_id=?, record_type=?, date=?, description=?, vet_name=?, cost=?, notes=? WHERE id=?`, e.AnimalId, e.RecordType, e.Date, e.Description, e.VetName, e.Cost, e.Notes, e.ID)
	return err
}

func (d *DB) DeleteHealthRecords(id string) error {
	_, err := d.db.Exec(`DELETE FROM health_records WHERE id=?`, id)
	return err
}

func (d *DB) CountHealthRecords() int {
	var n int; d.db.QueryRow(`SELECT COUNT(*) FROM health_records`).Scan(&n); return n
}

func (d *DB) SearchHealthRecords(q string, filters map[string]string) []HealthRecords {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (animal_id LIKE ? OR description LIKE ? OR vet_name LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["record_type"]; ok && v != "" { where += " AND record_type=?"; args = append(args, v) }
	rows, _ := d.db.Query(`SELECT id, animal_id, record_type, date, description, vet_name, cost, notes, created_at FROM health_records WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil { return nil }; defer rows.Close()
	var o []HealthRecords
	for rows.Next() { var e HealthRecords; rows.Scan(&e.ID, &e.AnimalId, &e.RecordType, &e.Date, &e.Description, &e.VetName, &e.Cost, &e.Notes, &e.CreatedAt); o = append(o, e) }
	return o
}
