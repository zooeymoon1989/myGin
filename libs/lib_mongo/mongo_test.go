// author: s0nnet
// time: 2020-09-01
// desc:

package lib_mongo

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
	"time"
)

func TestMongoSession(t *testing.T) {
	const (
		User     = "user_adm"
		Host     = "127.0.0.1:27030"
		Password = "this_is_test_pass"
		DBName   = "db_test"
	)
	type test_mongo struct {
		ID         primitive.ObjectID `bson:"_id"`
		Name       string             `bson:"name"`
		Phone      string             `bson:"phone"`
		CreateTime time.Time          `bson:"create_time"`
	}

	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s/%s?authSource=admin", User, Password, Host, DBName)

	Convey("test lib_mongo adaptor", t, func() {
		ms := NewMongoSession()
		err := ms.Connect(mongoURL, DBName)
		So(err, ShouldBeNil)
		So(ms.dbName, ShouldEqual, "ecos")

		// 测试查询命令
		Convey("test find opt", func() {
			ID, err := primitive.ObjectIDFromHex("5c061dc04d0e5544c4b7d31d")

			//
			err = ms.session.DB(DBName).C("test_mongo_find").Insert(&test_mongo{
				ID:         ID,
				Name:       "test",
				Phone:      "18866662222",
				CreateTime: time.Now().Local(),
			})
			So(err, ShouldBeNil)

			tx := test_mongo{}
			err = ms.session.DB(DBName).C("test_mongo_find").Find(bson.M{"_id": ID}).One(&tx)
			So(err, ShouldBeNil)
			So(tx.Name, ShouldEqual, "test")

			count, err := ms.FindCount("test_mongo_find", bson.M{})
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 1)
		})
	})
}
