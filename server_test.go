package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/storage"
)

func TestPublish_SingleMessage(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		topic = "test_topic"
		value = "test_value"
	)

	mockBroker := NewMockbrokerer(ctrl)
	mockBroker.EXPECT().Publish(topic, []byte(value))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/publish/%s", topic), strings.NewReader(value))

	srv := newServer(mockBroker)
	srv.ServeHTTP(rec, req)

	assert.Equal(http.StatusOK, rec.Code)
}

func TestSubscribe_SingleMessage(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		topic = "test_topic"
		msg   = "test_message"
	)

	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	assert.NoError(err)

	b := newBroker(&store{db: db})

	// Publish to the topic
	pubW := NewRecorder()
	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/publish/%s", topic), strings.NewReader(msg))
	r = mux.SetURLVars(r, map[string]string{"topic": topic})

	publish(b)(pubW, r)
	assert.Equal(http.StatusOK, pubW.Code)

	// Subscribe to the same topic
	subW := NewRecorder()
	r = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/subscribe/%s", topic), nil)
	r = mux.SetURLVars(r, map[string]string{"topic": topic})

	go subscribe(b)(subW, r)

	// Wait for the first message to be written
	decoder := NewDecodeWaiter(subW)

	// Read the first message
	var out string
	assert.NoError(decoder.WaitAndDecode(&out))
	assert.Equal(msg, out)
}

func TestSubscribe_WithAck(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		topic = "test_topic"
		msg   = "test_message"
	)

	db, err := leveldb.Open(storage.NewMemStorage(), nil)
	assert.NoError(err)

	b := newBroker(&store{db: db})

	// Publish to the topic
	pubW := NewRecorder()
	r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/publish/%s", topic), strings.NewReader(msg))
	r = mux.SetURLVars(r, map[string]string{"topic": topic})

	// Publish twice
	publish(b)(pubW, r)
	assert.Equal(http.StatusOK, pubW.Code)

	publish(b)(pubW, r)
	assert.Equal(http.StatusOK, pubW.Code)

	// Subscribe to the same topic
	subW := NewRecorder()
	r = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/subscribe/%s", topic), nil)
	r = mux.SetURLVars(r, map[string]string{"topic": topic})

	go subscribe(b)(subW, r)

	// Wait for the first message to be written
	decoder := NewDecodeWaiter(subW)

	// Read the first message
	var out string
	assert.NoError(decoder.WaitAndDecode(&out))
	assert.Equal(msg, out)
}
