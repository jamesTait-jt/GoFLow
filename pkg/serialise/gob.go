package serialise

import (
	"bytes"
	"encoding/gob"
)

type GobSerialiser[T any] struct{}

func NewGobSerialiser[T any]() *GobSerialiser[T] {
	return &GobSerialiser[T]{}
}

func (s *GobSerialiser[T]) Serialise(t T) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(t)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *GobSerialiser[T]) Deserialise(data []byte) (T, error) {
	var t T

	buf := bytes.NewBuffer(data)

	decoder := gob.NewDecoder(buf)

	err := decoder.Decode(&t)

	if err != nil {
		return t, err
	}

	return t, nil
}
