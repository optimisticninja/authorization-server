package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/optimisticninja/auth/logger"
	"github.com/optimisticninja/osin"
	"github.com/optimisticninja/osin-postgres/storage/postgres"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/sirupsen/logrus"
)

var (
	log        = logger.GetLogger(logrus.Fields{})
	dbUser     = getEnvVar("DB_USER")
	dbPassword = getEnvVar("DB_PASSWORD")
	dbHost     = getEnvVar("DB_HOST")
	// FIXME: Do lookups elsewhere and panic on missing values
	dbPort, _ = strconv.Atoi(getEnvVar("DB_PORT"))
	db        = getEnvVar("DB")

	certPath   = getEnvVar("CERT_PATH")
	keyPath    = getEnvVar("KEY_PATH")
	serverHost = getEnvVar("SERVER_HOST") //":14000"
)

func getEnvVar(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}
	return os.Getenv(key)
}

func HandleLoginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	if r.Method == "POST" && r.FormValue("login") == "test" && r.FormValue("password") == "test" {
		return true
	}

	w.Write([]byte("<html><body>"))

	w.Write([]byte(fmt.Sprintf("LOGIN %s (use test/test)<br/>", ar.Client.GetId())))
	w.Write([]byte(fmt.Sprintf("<form action=\"/authorize?%s\" method=\"POST\">", r.URL.RawQuery)))

	w.Write([]byte("Login: <input type=\"text\" name=\"login\" /><br/>"))
	w.Write([]byte("Password: <input type=\"password\" name=\"password\" /><br/>"))
	w.Write([]byte("<input type=\"submit\"/>"))

	w.Write([]byte("</form>"))

	w.Write([]byte("</body></html>"))

	return false
}

func main() {
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
			if !HandleLoginPage(ar, w, r) {
				return
			}
			ar.UserData = "test"
			ar.Authorized = true
			scopes := make(map[string]bool)
			for _, s := range strings.Fields(ar.Scope) {
				scopes[s] = true
			}
			server.FinishAuthorizeRequest(resp, r, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			log.Errorln(resp.InternalError)
		}
		//if !resp.IsError {
		//	resp.Output["custom_parameter"] = 187723
		//}
		osin.OutputJSON(resp, w, r)
		//resp := server.NewResponse()
		//defer resp.Close()

		//if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {

		//	// HANDLE LOGIN PAGE HERE

		//	ar.Authorized = true
		//	server.FinishAuthorizeRequest(resp, r, ar)
		//}
		//osin.OutputJSON(resp, w, r)
	})

	// Access token endpoint
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		//resp := server.NewResponse()
		//defer resp.Close()

		//if ar := server.HandleAccessRequest(resp, r); ar != nil {
		//	ar.Authorized = true
		//	server.FinishAccessRequest(resp, r, ar)
		//}
		//osin.OutputJSON(resp, w, r)
		resp := server.NewResponse()
		defer resp.Close()

		if ar := server.HandleAccessRequest(resp, r); ar != nil {
			switch ar.Type {
			case osin.AUTHORIZATION_CODE:
				ar.Authorized = true
			case osin.REFRESH_TOKEN:
				ar.Authorized = true
			case osin.PASSWORD:
				if ar.Username == "test" && ar.Password == "test" {
					ar.Authorized = true
				}
			case osin.CLIENT_CREDENTIALS:
				ar.Authorized = true
			case osin.ASSERTION:
				if ar.AssertionType == "urn:osin.example.complete" && ar.Assertion == "osin.data" {
					ar.Authorized = true
				}
			}
			server.FinishAccessRequest(resp, r, ar)
		}
		if resp.IsError && resp.InternalError != nil {
			log.Errorln(resp.InternalError)
		}
		//if !resp.IsError {
		//	resp.Output["custom_parameter"] = 19923
		//}
		osin.OutputJSON(resp, w, r)
	})

	// Information endpoint
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		resp := server.NewResponse()
		defer resp.Close()

		if ir := server.HandleInfoRequest(resp, r); ir != nil {
			server.FinishInfoRequest(resp, r, ir)
		}
		osin.OutputJSON(resp, w, r)
	})

	if *tcp {
		log.Printf("Listening on UDP/TCP...")
		http3.ListenAndServe(serverHost, certPath, keyPath, logger.RequestLoggerMiddleware(mux))
	} else {
		quicConf := &quic.Config{}
		s := http3.Server{
			Handler:    logger.RequestLoggerMiddleware(mux),
			Addr:       serverHost,
			QuicConfig: quicConf,
		}
		log.Printf("Listening on UDP...")
		err = s.ListenAndServeTLS(certPath, keyPath)
		if err != nil {
			log.Errorln(err)
		}
	}
}
