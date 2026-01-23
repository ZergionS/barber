package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

// ================= STRUCT =================
type Order struct {
	ID               int
	NamaPelanggan    string
	JarakKm          float64
	ModelRambut      string
	Harga            int
	PerkiraanWaktu   int
	MetodePembayaran string
	BuktiPembayaran  string
	CreatedAt        string
}

// ================= WEB HANDLER =================
func webHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// POST dari form HTML
		if r.Method == http.MethodPost {
			nama := r.FormValue("nama_pelanggan")
			jarak, _ := strconv.ParseFloat(r.FormValue("jarak_km"), 64)
			model := r.FormValue("model_rambut")
			metode := r.FormValue("metode_pembayaran")

			harga := 25000
			switch model {
			case "fade":
				harga = 30000
			case "undercut":
				harga = 35000
			case "gundul":
				harga = 20000
			}

			perkiraanWaktu := int(jarak * 30)

			_, err := db.Exec(`
				INSERT INTO orders 
				(nama_pelanggan, jarak_km, model_rambut, harga, perkiraan_waktu, metode_pembayaran)
				VALUES (?, ?, ?, ?, ?, ?)`,
				nama, jarak, model, harga, perkiraanWaktu, metode,
			)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// GET tampilkan data
		rows, err := db.Query("SELECT * FROM orders")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		orders := []Order{}

		for rows.Next() {
			var o Order
			var bukti, created sql.NullString

			rows.Scan(
				&o.ID,
				&o.NamaPelanggan,
				&o.JarakKm,
				&o.ModelRambut,
				&o.Harga,
				&o.PerkiraanWaktu,
				&o.MetodePembayaran,
				&bukti,
				&created,
			)

			if bukti.Valid {
				o.BuktiPembayaran = bukti.String
			}
			if created.Valid {
				o.CreatedAt = created.String
			}

			orders = append(orders, o)
		}

		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.Execute(w, orders)
	}
}

// ================= API HANDLER =================
func ordersAPI(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {

		// ===== READ =====
		case http.MethodGet:
			rows, _ := db.Query("SELECT * FROM orders")
			defer rows.Close()

			orders := []Order{}
			for rows.Next() {
				var o Order
				var bukti, created sql.NullString

				rows.Scan(
					&o.ID,
					&o.NamaPelanggan,
					&o.JarakKm,
					&o.ModelRambut,
					&o.Harga,
					&o.PerkiraanWaktu,
					&o.MetodePembayaran,
					&bukti,
					&created,
				)

				if bukti.Valid {
					o.BuktiPembayaran = bukti.String
				}
				if created.Valid {
					o.CreatedAt = created.String
				}

				orders = append(orders, o)
			}

			json.NewEncoder(w).Encode(orders)

		// ===== CREATE =====
		case http.MethodPost:
			var input struct {
				NamaPelanggan    string  `json:"nama_pelanggan"`
				JarakKm          float64 `json:"jarak_km"`
				ModelRambut      string  `json:"model_rambut"`
				MetodePembayaran string  `json:"metode_pembayaran"`
			}

			json.NewDecoder(r.Body).Decode(&input)

			harga := 25000
			switch input.ModelRambut {
			case "fade":
				harga = 30000
			case "undercut":
				harga = 35000
			case "gundul":
				harga = 20000
			}

			perkiraanWaktu := int(input.JarakKm * 30)

			db.Exec(`
				INSERT INTO orders 
				(nama_pelanggan, jarak_km, model_rambut, harga, perkiraan_waktu, metode_pembayaran)
				VALUES (?, ?, ?, ?, ?, ?)`,
				input.NamaPelanggan,
				input.JarakKm,
				input.ModelRambut,
				harga,
				perkiraanWaktu,
				input.MetodePembayaran,
			)

			json.NewEncoder(w).Encode(map[string]string{
				"message": "order berhasil dibuat",
			})

		// ===== UPDATE =====
		case http.MethodPut:
			idStr := r.URL.Query().Get("id")
			id, _ := strconv.Atoi(idStr)

			var input struct {
				NamaPelanggan    string  `json:"nama_pelanggan"`
				JarakKm          float64 `json:"jarak_km"`
				ModelRambut      string  `json:"model_rambut"`
				MetodePembayaran string  `json:"metode_pembayaran"`
			}
			json.NewDecoder(r.Body).Decode(&input)

			harga := 25000
			switch input.ModelRambut {
			case "fade":
				harga = 30000
			case "undercut":
				harga = 35000
			case "gundul":
				harga = 20000
			}
			perkiraanWaktu := int(input.JarakKm * 30)

			_, err := db.Exec(`
				UPDATE orders
				SET nama_pelanggan=?, jarak_km=?, model_rambut=?, harga=?, perkiraan_waktu=?, metode_pembayaran=?
				WHERE id=?`,
				input.NamaPelanggan, input.JarakKm, input.ModelRambut, harga, perkiraanWaktu, input.MetodePembayaran, id,
			)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "order berhasil diupdate"})

		// ===== DELETE =====
		case http.MethodDelete:
			idStr := r.URL.Query().Get("id")
			id, _ := strconv.Atoi(idStr)

			_, err := db.Exec(`DELETE FROM orders WHERE id=?`, id)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "order berhasil dihapus"})

		default:
			http.Error(w, "Method not allowed", 405)
		}
	}
}

// ================= MAIN =================
func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/cukur_panggilan"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// ROUTES
	http.HandleFunc("/", webHandler(db))
	http.HandleFunc("/orders", ordersAPI(db))

	// STATIC FILE SERVER (WAJIB UNTUK GAMBAR)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Server jalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
