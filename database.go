package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps               map[int]Chirp        `json:"chirps"`
	Users                map[int]User         `json:"users"`
	EmailIDUserMap       map[string]int       `json:"emailIDUserMap"`
	RevokedRefreshTokens map[string]time.Time `json:"revokedRefreshTokens"`
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps:               map[int]Chirp{},
		Users:                map[int]User{},
		EmailIDUserMap:       map[string]int{},
		RevokedRefreshTokens: map[string]time.Time{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) CreateUser(email string, pwd string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	if _, ok := dbStructure.EmailIDUserMap[email]; ok {
		return User{}, fmt.Errorf("user with email %s already exists", email)
	}

	id := len(dbStructure.Users) + 1
	user := User{
		ID:          id,
		Email:       email,
		Password:    string(pwd),
		IsChirpyRed: false,
	}
	dbStructure.Users[id] = user
	dbStructure.EmailIDUserMap[email] = id
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpdateUser(id int, email string, pwd string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	u, ok := dbStructure.Users[id]
	if !ok {
		return User{}, fmt.Errorf("user does not exists: %v", id)
	}
	oldEmail := u.Email
	delete(dbStructure.EmailIDUserMap, oldEmail)
	user := User{
		ID:          id,
		Email:       email,
		Password:    string(pwd),
		IsChirpyRed: u.IsChirpyRed,
	}
	dbStructure.Users[id] = user
	dbStructure.EmailIDUserMap[email] = id
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpgradeUser(id int) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}

	u, ok := dbStructure.Users[id]
	if !ok {
		return false, fmt.Errorf("user does not exists: %v", id)
	}
	user := User{
		ID:          id,
		Email:       u.Email,
		Password:    u.Password,
		IsChirpyRed: true,
	}
	dbStructure.Users[id] = user
	err = db.writeDB(dbStructure)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *DB) RevokeRefreshToken(refreshToken string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}
	dbStructure.RevokedRefreshTokens[refreshToken] = time.Now().UTC()
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CheckRefreshToken(refreshToken string) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}
	_, ok := dbStructure.RevokedRefreshTokens[refreshToken]
	return ok, nil
}

func (db *DB) FindUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id, ok := dbStructure.EmailIDUserMap[email]
	if !ok {
		return User{}, errors.New("user does not exists")
	}
	user := dbStructure.Users[id]
	return user, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, authorId int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:       id,
		Body:     body,
		AuthorId: authorId,
	}
	dbStructure.Chirps[id] = chirp
	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps(authorIds ...int) ([]Chirp, error) {
	authorId := 0
	if len(authorIds) > 0 {
		authorId = authorIds[0]
	}
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		if authorId == 0 {
			chirps = append(chirps, chirp)
		} else if authorId == chirp.AuthorId {
			chirps = append(chirps, chirp)
		}

	}

	return chirps, nil
}

func (db *DB) GetChirp(ID int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	if chirp, ok := dbStructure.Chirps[ID]; ok {
		fmt.Printf("Chirp %v found, author id: %v\n", ID, chirp.AuthorId)
		return chirp, nil
	}

	return Chirp{}, errors.New(fmt.Sprintf("unable to find chirp id %v", ID))
}

func (db *DB) DeleteChirp(ID int, UserId int) (bool, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return false, err
	}

	chirp, ok := dbStructure.Chirps[ID]
	if !ok {
		return false, errors.New(fmt.Sprintf("unable to find chirp id %v", ID))
	}
	if chirp.AuthorId != UserId {
		return false, errors.New(fmt.Sprintf("unauthorized to perform task", ID))
	}
	delete(dbStructure.Chirps, ID)

	err = db.writeDB(dbStructure)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func DeleteDB(fileName string) {
	_, err := os.ReadFile(fileName)
	if err == nil {
		err := os.Remove(fileName)
		if err != nil {
			fmt.Printf("error deleting DB %s: %v", fileName, err.Error())
		}
	}
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}
	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}
	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}
