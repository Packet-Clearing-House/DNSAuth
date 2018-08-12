package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/elgs/gosqljson"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// http://localhost:8080/customers/list
func getCustomerList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		db, err := sql.Open("postgres", "postgres://user@127.0.0.1/pipeline?sslmode=disable")
		if err != nil {
			log.Println(err)
		}
		data, _ := gosqljson.QueryDbToMap(db, "lower", "SELECT ip, name, asn, prefix FROM ns_customers;")
		//log.Println(headers)
		//log.Println(data)
		json, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}
		w.Write([]byte("{\"data\":"))
		w.Write(json)
		w.Write([]byte("}"))
	}

}

//http://localhost:8080/customers/add?name=test&ip=216.21.3.3&prefix=false&asn=false
func addCustomer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		name := r.URL.Query().Get("name")
		ip := r.URL.Query().Get("ip")
		prefix := r.URL.Query().Get("prefix")
		asn := r.URL.Query().Get("asn")

		db, err := sql.Open("postgres", "postgres://user@127.0.0.1/pipeline?sslmode=disable")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			_, err := db.Query("INSERT INTO ns_customers VALUES ($1, $2, $3, $4);", ip, name, prefix, asn)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}
	}
}

// http://localhost:8080/customers/remove?ip=216.21.3.3
func removeCustomer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ip := r.URL.Query().Get("ip")
		if ip == "" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("No IP provided to delete customer!"))
		}

		db, err := sql.Open("postgres", "postgres://user@127.0.0.1/pipeline?sslmode=disable")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			_, err := db.Query("DELETE FROM ns_customers WHERE ip = $1;", ip)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}
	}
}

// http://localhost:8080/customers/update?name=foo&ip=1.199.71.1&prefix=false&asn=false
func updateCustomer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		name := r.URL.Query().Get("name")
		ip := r.URL.Query().Get("ip")
		prefix := r.URL.Query().Get("prefix")
		asn := r.URL.Query().Get("asn")

		db, err := sql.Open("postgres", "postgres://user@127.0.0.1/pipeline?sslmode=disable")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		} else {
			_, err := db.Query("UPDATE ns_customers SET name = $1, prefix = $2, asn = $3 WHERE ip = $4;", name, prefix, asn, ip)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
		}
	}
}

func main() {

	mux := http.NewServeMux()

	mux.Handle("/customers/list", getCustomerList())
	mux.Handle("/customers/add", addCustomer())
	mux.Handle("/customers/delete", removeCustomer())
	mux.Handle("/customers/update", updateCustomer())

	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handler)

}
