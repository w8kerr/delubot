package mongo

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/globalsign/mgo"
	"github.com/sirupsen/logrus"
)

//MDB is the global mongo session do not use it directly, always call .Clone() then defer sess.Close()
var MDB *mgo.Session

var MONGO_CONNECTION string
var DB_NAME string

//Init connect to mongo db and initialize MDB
func Init(quiet bool) error {
	MONGO_CONNECTION = os.Getenv("MONGO_CONNECTION")
	DB_NAME = os.Getenv("DB_NAME")

	if DB_NAME == "" {
		DB_NAME = "deluclub-dev"
	}

	mgo.SetDebug(false)
	if !quiet {
		mgo.SetLogger(log.New(os.Stderr, "MONGO:", 1))
	}

	//If we have a prod url set, use the SSL connection method
	if MONGO_CONNECTION != "" {
		tlsConfig := &tls.Config{}
		tlsConfig.InsecureSkipVerify = true
		dialInfo, err := mgo.ParseURL(MONGO_CONNECTION)
		if err != nil {
			logrus.WithField("err", err).Error("Failed to parse mongo url")
			return err
		}

		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
		//Here it is the session. Up to you from here ;)
		session, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			logrus.WithField("err", err).Error("Unable to connect to mongo")
			return err
		}
		MDB = session
		MDB.SetPoolLimit(20)
	} else {
		session, err := mgo.Dial("")
		if err != nil {
			logrus.WithField("err", err).Error("Unable to connect to mongo")
			return err
		}
		MDB = session
		MDB.SetPoolLimit(20)
	}

	MDB.SetMode(mgo.Strong, true)

	ensureIndexes()

	//Check if we have any data, if not create a default client and admin user
	//db := MDB.DB(options.CENTRAL_DB)
	//c, _ := db.C("clients").Count()
	//if c == 0 {
	//	createDefaultClient(db)
	//}

	return nil
}

// CaseInsensitive A collation that makes case insensitive comparisons
var CaseInsensitive = &mgo.Collation{
	Locale:   "en",
	Strength: 1,
}

func ensureIndexes() {
	//common db
	fmt.Println("INDEXING:", DB_NAME)
}

func createNormalIndex(collection string, index []string) {
	idx := mgo.Index{
		Key:        index,
		Unique:     false,
		Background: true,
		DropDups:   false,
	}
	err := MDB.DB(DB_NAME).C(collection).EnsureIndex(idx)
	if err != nil {
		panic(err)
	}
}

func createUniqueIndex(collection string, index []string) {
	idx := mgo.Index{
		Key:        index,
		Unique:     true,
		Background: true,
		DropDups:   false,
		Sparse:     true,
	}
	err := MDB.DB(DB_NAME).C(collection).EnsureIndex(idx)
	if err != nil {
		panic(err)
	}
}

//CentralDB returns the Database with the configured properties
func CentralDB(session *mgo.Session) *mgo.Database {
	return session.DB(DB_NAME)
}
