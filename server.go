package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/optimisticninja/osin-postgres/storage/postgres"
	_ "github.com/lib/pq"
	"github.com/optimisticninja/osin"
	"github.com/optimisticninja/auth/logger"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

const (
	dbUser = "root"
	dbPassword = "root"
	dbHost = "127.0.0.1"
	dbPort = 5432
	db = "auth"

	certPath = "./cert/cert.pem"
	keyPath = "./cert/priv.key"
	serverHost = ":14000"
)

func main() {
	log := logger.GetLogger(logrus.Fields{})
	tcp := flag.Bool("tcp", false, "also listen on TCP")
	flag.Parse()

	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable", dbUser, db, dbPassword, dbHost)
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully connected to [%s]", connStr)

	store := postgres.New(db)
	config := osin.NewServerConfig()
	config.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	config.AllowedAccessTypes = osin.AllowedAccessType{
		osin.AUTHORIZATION_CODE,
		osin.REFRESH_TOKEN, 
		osin.PASSWORD, 
		osin.CLIENT_CREDENTIALS, 
		osin.ASSERTION}
	config.AllowGetAccessRequest = true
	config.AllowClientSecretInParams = true
	server := osin.NewServer(config, store)

	mux := http.NewServeMux()
	// Authorization code endpoint
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()

		if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {

			// HANDLE LOGIN PAGE HERE

			ar.Authorized = true
			server.FinishAuthorizeRequest(resp, r, ar)
		}
		osin.OutputJSON(resp, w, r)
	})

	// Access token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()

		if ar := server.HandleAccessRequest(resp, r); ar != nil {
			ar.Authorized = true
			server.FinishAccessRequest(resp, r, ar)
		}
		osin.OutputJSON(resp, w, r)
	})

	if *tcp {
		log.Printf("Listening on UDP/TCP...")
		http3.ListenAndServe(":14000", "./cert/cert.pem", "./cert/priv.key", logger.RequestLoggerMiddleware(mux))
	} else {
		quicConf := &quic.Config{}
		s := http3.Server{
			Handler:    logger.RequestLoggerMiddleware(mux),
			Addr:       ":14000",
			QuicConfig: quicConf,
		}
		log.Printf("Listening on UDP...")
		err = s.ListenAndServeTLS(certPath, keyPath)
		if err != nil {
			fmt.Println(err)
		}
	}
}
