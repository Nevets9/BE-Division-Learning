package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"tugas5/database"
	"tugas5/models"
	"tugas5/utils"

	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	owner := r.FormValue("owner")
	name := r.FormValue("name")
	password := r.FormValue("password")

	if(owner == "" || name == "" || password == ""){
		http.Error(w, "owner, name, or password can not be empty", http.StatusInternalServerError)
		return
	}


	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec("INSERT INTO stores (owner, password, name) VALUES (?, ?, ?)", owner, hash, name)
	if err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	owner := r.FormValue("owner")
	password := r.FormValue("password")

	var store models.Store
	err := database.DB.QueryRow("SELECT id, owner, password, token FROM stores WHERE owner = ? AND deleted_at IS NULL", owner).
		Scan(&store.ID, &store.Owner, &store.Password, &store.Token)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(store.Password), []byte(password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := utils.GenerateToken(32)

	_, err = database.DB.Exec("UPDATE stores SET token = ? WHERE owner = ? AND deleted_at IS NULL", token, owner)
	if err != nil {
		http.Error(w, "Failed to set token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}


func AuthStore(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	owner := r.FormValue("owner")
	password := r.FormValue("password")

	if owner == "" || password == "" {
		http.Error(w, "owner or password cannot be empty", http.StatusBadRequest)
		return
	}

	var store models.Store
	err = database.DB.QueryRow("SELECT id, name, owner, password FROM stores WHERE owner = ? AND password = ? AND deleted_at IS NULL", owner, password).
		Scan(&store.ID, &store.Name, &store.Owner, &store.Password)
	if err != nil {
		http.Error(w, "owner or password is wrong", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login Success",
		"store":   store,
	})
}

func GetStores(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, name, owner, password, token, role, deleted_at FROM stores WHERE deleted_at IS NULL")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	stores := []models.Store{}
	for rows.Next() {
		store := models.Store{}
		err := rows.Scan(
			&store.ID,
			&store.Name,
			&store.Owner,
			&store.Password,
			&store.Token,
			&store.Role,
			&store.DeletedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		stores = append(stores, store)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Success",
		"stores":  stores,
	})
}

func GetStore(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		http.Error(w, "ID cannot be empty", http.StatusBadRequest)
		return
	}

	var store models.Store
	err := database.DB.QueryRow(`
		SELECT id, name, owner, token, role, deleted_at
		FROM stores WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(
		&store.ID,
		&store.Name,
		&store.Owner,
		&store.Token,
		&store.Role,
		&store.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Store found",
		"store":   store,
	})
}

func UpdateStore(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		http.Error(w, "ID cannot be empty", http.StatusBadRequest)
		return
	}

	var store models.Store
	err := database.DB.QueryRow("SELECT id, name, owner, role FROM stores WHERE id = ? AND deleted_at IS NULL", id).
		Scan(&store.ID, &store.Name, &store.Owner, &store.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Store not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	owner := r.FormValue("owner")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if name != "" {
		store.Name = name
	}
	if owner != "" {
		store.Owner = owner
	}
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		store.Password = string(hashedPassword)
		_, err = database.DB.Exec(
			"UPDATE stores SET name = ?, owner = ?, password = ?, role = ? WHERE id = ? AND deleted_at IS NULL",
			store.Name, store.Owner, store.Password, role, store.ID,
		)
	} else {
		if role == "" {
			role = "user"
		}

		_, err = database.DB.Exec(
			"UPDATE stores SET name = ?, owner = ?, role = ? WHERE id = ? AND deleted_at IS NULL",
			store.Name, store.Owner, role, store.ID,
		)
	}

	if err != nil {
		http.Error(w, "Failed to update store: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Store updated",
		"store": map[string]interface{}{
			"id":    store.ID,
			"name":  store.Name,
			"owner": store.Owner,
			"role":  role, 
		},
	})
}

func DeleteStore(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {		
		http.Error(w, "id is null", http.StatusBadRequest)
		return
	}

	if id == "" {
		http.Error(w, "id is null", http.StatusBadRequest)
		return
	}

	var exists bool
	err := database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM stores WHERE id = ? AND deleted_at IS NULL)", id).Scan(&exists)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "store not found", http.StatusNotFound)
		return
	}

	_, err = database.DB.Exec("UPDATE stores SET deleted_at = NOW() WHERE id = ?", id)
	if err != nil {
		http.Error(w, "failed to delete store", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Store Deleted",
		"id":      id,
	})
}