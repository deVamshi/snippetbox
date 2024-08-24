package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/deVamshi/snippetbox/internal/models"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// instead of global variables we define like this
// incase our handlers are in different packages this won't work
// instead need to use closure approach or middlewares
type application struct {
	infoLogger     *log.Logger
	errorLogger    *log.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {

	// flags
	addr := flag.String("addr", ":4000", "port")
	dsn := flag.String("dsn", "web:fass@/snippetbox?parseTime=true", "MySQL data source name")

	flag.Parse()

	infoLogger := log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
	errorLogger := log.New(os.Stderr, "ERROR:\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(*dsn)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLogger.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = time.Hour * 12
	sessionManager.Cookie.Secure = true

	app := &application{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		snippets: &models.SnippetModel{
			DB: db,
		},
		templateCache:  templateCache,
		formDecoder:    form.NewDecoder(),
		sessionManager: sessionManager,
	}

	tlsCfg := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := http.Server{
		Addr:      *addr,
		ErrorLog:  app.errorLogger,
		Handler:   app.routes(),
		TLSConfig: tlsCfg,

		// after what time the keep-alive connections
		// need to be closed
		IdleTimeout: time.Minute,
		// time duration within which the headers and body
		// the server should read
		ReadTimeout: time.Second * 5,
		// time duration within which the server should take
		// to write as response
		WriteTimeout: time.Second * 10,
	}

	app.infoLogger.Printf("Listening on https://localhost%s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	app.errorLogger.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// 426/ 252
