# DynamoStore [![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/denizeren/dynamostore) [![Build Status](https://travis-ci.org/denizeren/dynamostore.svg?branch=master)](https://travis-ci.org/denizeren/dynamostore)

A session store backend for [gorilla/sessions](http://www.gorillatoolkit.org/pkg/sessions) - [src](https://github.com/gorilla/sessions).

## Requirements

Depends on the [Goamz/aws](https://github.com/crowdmob/goamz/aws) Go Amazon Library

Depends on the [Goamz/dynamodb](https://github.com/crowdmob/goamz/dynamodb) Go Amazon Dynamodb Library

## Installation

    go get github.com/denizeren/dynamostore

## Documentation

Available on [godoc.org](http://godoc.org/github.com/denizeren/dynamostore).

See http://www.gorillatoolkit.org/pkg/sessions for full documentation on underlying interface.

### Example

    // Fetch new store.
    store, err := NewDynamoStore("AWS_ACCESS_KEY", "AWS_SECRET_KEY", "DYNAMODB_TABLE_NAME", "AWS_REGION_NAME", "SECRET-KEY")
    if err != nil {
        panic(err)
    }

    // Get a session.
    session, err = store.Get(req, "session-key")
    if err != nil {
        log.Error(err.Error())
    }

    // Add a value.
    session.Values["foo"] = "bar"

    // Save.
    if err = sessions.Save(req, rsp); err != nil {
        t.Fatalf("Error saving session: %v", err)
    }

    // Delete session.
    session.Options.MaxAge = -1
    if err = sessions.Save(req, rsp); err != nil {
        t.Fatalf("Error saving session: %v", err)
    }
