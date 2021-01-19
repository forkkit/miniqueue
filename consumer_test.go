package main

import (
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestConsumerNext(t *testing.T) {
	assert := assert.New(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		topic = "test_topic"
		msg1  = []byte("message1")
		msg2  = []byte("message2")
	)

	mockStore := NewMockstorer(ctrl)
	mockStore.EXPECT().GetNext(topic).Return(msg1, nil)
	mockStore.EXPECT().GetNext(topic).Return(msg2, nil)

	b := newBroker(mockStore)
	c := b.Subscribe(topic)

	val, err := c.Next()
	assert.NoError(err)
	assert.Equal(msg1, val)

	val, err = c.Next()
	assert.NoError(err)
	assert.Equal(msg2, val)
}
