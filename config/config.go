/*
 * Realtime config from the DB
 */

package config

import (
	"context"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/w8kerr/delubot/models"
	"github.com/w8kerr/delubot/mongo"
)

// Get Load the config object
func Get(c context.Context) models.Config {
	session := mongo.MDB.Clone()
	session.SetMode(mgo.Strong, false)
	db := session.DB(mongo.DB_NAME)

	configCol := db.C("config")
	config := models.Config{}
	configCol.Find(bson.M{}).One(&config)

	return config
}
