// Coder: koalatea
// Email: koalateac@gmail.com

package main

import (
	"flag"
	"log"
	"time"

	"github.com/alexedwards/scs"
	"github.com/alexedwards/scs/stores/mysqlstore"
	"github.com/juju/loggo"
	"github.com/koalatea/changan/pkg/models"
)

func init() {
}

func main() {
	logger := loggo.GetLogger("changan")
	logger.SetLogLevel(loggo.DEBUG)
	logger.Debugf("Booting ChangAn")

	addr := flag.String("addr", ":8080", "HTTP Network Address and port")
	dsn := flag.String("dsn", "root:@/changan_test", "MySQL DSN")
	databaseName := flag.String("database-name", "changan_test", "MongoDBName")
	htmlDir := flag.String("html-dir", "./ui/html", "Path to HTML templates")
	staticDir := flag.String("static-dir", "./ui/static", "Path to static Directory")
	tlsCert := flag.String("tls-cert", "./tls/cert.pem", "Path to TLS certificate")
	tlsKey := flag.String("tls-key", "./tls/key.pem", "Path to TLS key")
	flag.Parse()

	dbSQL := openDB(*dsn) // need for sessions at the moment.
	defer dbSQL.Close()
	sqlx := models.OpenMysqlDB(*dsn)
	defer sqlx.Close()
	mongoSession, mongoDB, err := models.OpenMongo(*databaseName)
	if err != nil {
		log.Fatal(err)
	}
	defer mongoSession.Close()

	//sessionManager := scs.NewCookieManager("abcd") //temporary will redo this to use mysql later TODO
	sessionManager := scs.NewManager(mysqlstore.New(dbSQL, 12*time.Hour))
	sessionManager.Lifetime(12 * time.Hour)
	sessionManager.Persist(true)

	app := &App{
		Addr:      *addr,
		Mongo:     &models.Database{mongoDB}, //TODO verify this works
		Database:  &models.SQLDatabase{sqlx},
		HTMLDir:   *htmlDir,
		StaticDir: *staticDir,
		Sessions:  sessionManager,
		TLSKey:    *tlsKey,
		TLSCert:   *tlsCert,
		Logger:    logger,
	}

	app.RunServer()
}
