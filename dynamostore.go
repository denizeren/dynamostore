// Copyright 2015 Deniz Eren. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package dynamostore

import (
	"bytes"
	"encoding/base32"
	"encoding/gob"
	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/dynamodb"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

type DynamoStore struct {
	Table   *dynamodb.Table
	Codecs  []securecookie.Codec
	Options *sessions.Options // default configuration
}

type DynamoData struct {
	Id   string
	Data []byte
}

func NewDynamoStore(accessKey string, secretKey string, tableName string, region string, keyPairs ...[]byte) (*DynamoStore, error) {
	regionObj := aws.GetRegion(region)
	return NewDynamoStoreWithRegionObj(accessKey, secretKey, tableName, regionObj, keyPairs...)
}

func NewDynamoStoreWithRegionObj(accessKey string, secretKey string, tableName string, region aws.Region, keyPairs ...[]byte) (*DynamoStore, error) {
	awsAuth := aws.Auth{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}

	server := dynamodb.New(awsAuth, region)
	table := server.NewTable(tableName, dynamodb.PrimaryKey{KeyAttribute: dynamodb.NewStringAttribute("Id", "")})

	dynStore := &DynamoStore{
		Table:  table,
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
	}

	return dynStore, nil
}

// Get returns a session for the given name after adding it to the registry.
//
// See gorilla/sessions FilesystemStore.Get().
// or  boj/redistore
func (s *DynamoStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *DynamoStore) New(r *http.Request, name string) (*sessions.Session, error) {
	var err error
	session := sessions.NewSession(s, name)

	// make a copy
	options := sessions.Options{}
	session.Options = &options
	session.IsNew = true
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err := s.load(session)
			if err == nil {
				session.IsNew = false
			} else {
				session.IsNew = true
			}
		}
	}

	return session, err
}

// Save adds a single session to the response.
func (s *DynamoStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Marked for deletion.
	if session.Options.MaxAge < 0 {
		if err := s.delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	} else {
		// Build an alphanumeric key for the redis store.
		if session.ID == "" {
			session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
		}
		if err := s.save(session); err != nil {
			return err
		}
		encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	}
	return nil
}

// save stores the session in redis.
func (s *DynamoStore) save(session *sessions.Session) error {
	// TODO expiration date to be added
	// TODO max length to be added
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(session.Values)
	if err != nil {
		return err
	}
	b := buf.Bytes()

	data := &DynamoData{Id: session.ID, Data: b}
	err = s.Table.PutDocument(&dynamodb.Key{HashKey: session.ID}, data)
	if err != nil {
		return err
	}

	return nil
}

// load reads the session from dynamodb.
// returns error if session data does not exist in dynamodb
func (s *DynamoStore) load(session *sessions.Session) error {
	var sessionData DynamoData
	err := s.Table.GetDocument(&dynamodb.Key{HashKey: session.ID}, &sessionData)
	if err != nil {
		return err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(sessionData.Data))
	return dec.Decode(&session.Values)
}

// delete removes keys from redis if MaxAge<0
func (s *DynamoStore) delete(session *sessions.Session) error {
	err := s.Table.DeleteDocument(&dynamodb.Key{HashKey: session.ID})
	if err != nil {
		return err
	}
	return nil
}
