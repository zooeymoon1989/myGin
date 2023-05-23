// author: s0nnet
// time: 2020-09-01
// desc:

package lib_mongo

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Session lib_mongo session
type Session struct {
	client      *mongo.Client
	collection  *mongo.Collection
	maxPoolSize uint64
	db          string
	uri         string
	m           sync.RWMutex
	filter      interface{}
	limit       *int64
	project     interface{}
	skip        *int64
	sort        interface{}
	distinct    interface{}
}

// New session
//
// Relevant documentation:
// https://docs.mongodb.com/manual/reference/connection-string/
func New(uri string) *Session {
	session := &Session{
		uri: uri,
	}
	return session
}

// SetDB set db
func (s *Session) SetDB(db string) {
	s.m.Lock()
	s.db = db
	s.m.Unlock()
}

// C Collection alias
func (s *Session) C(collection string) *Collection {
	if len(s.db) == 0 {
		s.db = "test"
	}
	d := &Database{database: s.client.Database(s.db)}
	return &Collection{collection: d.database.Collection(collection)}
}

// Collection returns collection
func (s *Session) Collection(collection string) *Collection {
	if len(s.db) == 0 {
		s.db = "test"
	}
	d := &Database{database: s.client.Database(s.db)}
	return &Collection{collection: d.database.Collection(collection)}
}

// SetPoolLimit specifies the max size of a server's connection pool.
func (s *Session) SetPoolLimit(limit uint64) {
	s.m.Lock()
	s.maxPoolSize = limit
	s.m.Unlock()
}

// Connect lib_mongo client
func (s *Session) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	opt := options.Client().ApplyURI(s.uri)
	opt.SetMaxPoolSize(s.maxPoolSize)

	client, err := mongo.NewClient(opt)
	if err != nil {
		return err
	}
	err = client.Connect(ctx)
	if err != nil {
		return err
	}
	s.client = client
	return nil
}

func (s *Session) Disconnect() {
	s.m.Lock()
	s.client.Disconnect(context.Background())
	s.m.Unlock()
}

// Ping verifies that the client can connect to the topology.
// If readPreference is nil then will use the client's default read
// preference.
func (s *Session) Ping() error {
	return s.client.Ping(context.TODO(), readpref.Primary())
}

// Client return lib_mongo Client
func (s *Session) Client() *mongo.Client {
	return s.client
}

// DB returns a value representing the named database.
func (s *Session) DB(db string) *Database {
	return &Database{database: s.client.Database(db)}
}

// Limit specifies a limit on the number of results.
// A negative limit implies that only 1 batch should be returned.
func (s *Session) Limit(limit int64) *Session {
	s.limit = &limit
	return s
}

// Skip specifies the number of documents to skip before returning.
// For server versions < 3.2, this defaults to 0.
func (s *Session) Skip(skip int64) *Session {
	s.skip = &skip
	return s
}

// Sort specifies the order in which to return documents.
func (s *Session) Sort(sort interface{}) *Session {
	s.sort = sort
	return s
}

// Select is used to determine which fields are displayed or not displayed in the returned results
// Format: bson.M{"age": 1} means that only the age field is displayed
func (s *Session) Select(projection interface{}) *Session {
	s.project = projection
	return s
}

// One returns one document
func (s *Session) One(result interface{}) error {
	opt := options.FindOne()

	if s.sort != nil {
		opt.SetSort(s.sort)
	}

	if s.project != nil {
		opt.SetProjection(s.project)
	}

	if s.skip != nil {
		opt.SetProjection(*s.skip)
	}

	data, err := s.collection.FindOne(context.TODO(), s.filter, opt).DecodeBytes()
	if err != nil {
		return err
	}
	err = bson.Unmarshal(data, result)
	return err
}

// All find all
func (s *Session) All(result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a %s", resultv.Kind())
	}
	slicev := resultv.Elem()

	if slicev.Kind() == reflect.Interface {
		slicev = slicev.Elem()
	}
	if slicev.Kind() != reflect.Slice {
		return fmt.Errorf("results argument must be a pointer to a slice, but was a pointer to %s", slicev.Kind())
	}

	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error

	opt := options.Find()

	if s.sort != nil {
		opt.SetSort(s.sort)
	}

	if s.project != nil {
		opt.SetProjection(s.project)
	}

	if s.limit != nil {
		opt.SetLimit(*s.limit)
	}

	if s.skip != nil {
		opt.SetSkip(*s.skip)
	}

	cur, err := s.collection.Find(ctx, s.filter, opt)
	defer cur.Close(ctx)
	if err != nil {
		return err
	}
	if err = cur.Err(); err != nil {
		return err
	}
	i := 0
	for cur.Next(ctx) {
		elemp := reflect.New(elemt)
		if err = bson.Unmarshal(cur.Current, elemp.Interface()); err != nil {
			return err
		}
		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}

// Pipe find all
func (s *Session) Pipe(pipeline, result interface{}) error {
	resultv := reflect.ValueOf(result)
	if resultv.Kind() != reflect.Ptr {
		panic("result argument must be a slice address")
	}
	slicev := resultv.Elem()

	if slicev.Kind() == reflect.Interface {
		slicev = slicev.Elem()
	}
	if slicev.Kind() != reflect.Slice {
		panic("result argument must be a slice address")
	}

	slicev = slicev.Slice(0, slicev.Cap())
	elemt := slicev.Type().Elem()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	opts := options.Aggregate()
	opts.SetAllowDiskUse(true)
	opts.SetBatchSize(5)

	cur, err := s.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)

	if err = cur.Err(); err != nil {
		return err
	}
	i := 0
	for cur.Next(ctx) {
		elemp := reflect.New(elemt)
		if err = bson.Unmarshal(cur.Current, elemp.Interface()); err != nil {
			return err
		}
		slicev = reflect.Append(slicev, elemp.Elem())
		i++
	}
	resultv.Elem().Set(slicev.Slice(0, i))
	return nil
}

func (s *Session) Distinct(distinct string) ([]interface{}, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	result, err := s.collection.Distinct(ctx, distinct, s.filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}
