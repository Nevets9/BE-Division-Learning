package routes

import (
	"tugas5/controllers"
	"tugas5/utils"
	"net/http"
)

func ProductsRoutes(){
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/products/", productsHandlerWithID)
}

func StoreRoutes(){
	http.HandleFunc("/store", storeHandler)
	http.HandleFunc("/store/", storeHandlerWithID)
	http.HandleFunc("/store/auth", storeAuthHandler)
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		controllers.GetProducts(w, r)
	case http.MethodPost:
		controllers.CreateProduct(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func productsHandlerWithID(w http.ResponseWriter, r *http.Request) {
	parts, err := utils.SplitPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := parts[2]
	switch r.Method {
	case http.MethodGet:
		controllers.GetProduct(w, r, id)
	case http.MethodPatch:
		controllers.UpdateProduct(w, r, id)
	case http.MethodDelete:
		controllers.DeleteProduct(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func storeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		controllers.GetStores(w, r)
	case http.MethodPost:
		controllers.CreateStore(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func storeHandlerWithID(w http.ResponseWriter, r *http.Request) {
	parts, err := utils.SplitPath(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	id := parts[2]
	switch r.Method {
	case http.MethodGet:
		controllers.GetStore(w, r, id)
	case http.MethodPatch:
		controllers.UpdateStore(w, r, id)
	case http.MethodDelete:
		controllers.DeleteStore(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func storeAuthHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		controllers.AuthStore(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}