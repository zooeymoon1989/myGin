// author: s0nnet
// time: 2020-09-01
// desc:

package lib_mongo

import "go.mongodb.org/mongo-driver/mongo"

type Database struct {
	database *mongo.Database
}

// returns collection
func (d *Database) C(collection string) *Collection {
	return &Collection{collection: d.database.Collection(collection)}
}

// returns collection
func (d *Database) Collection(collection string) *Collection {
	return &Collection{collection: d.database.Collection(collection)}
}
