package dao

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func SerializeFloat64(vector []float64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, vector)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize float64 slice: %w", err)
	}
	return buf.Bytes(), nil
}

func DeserializeFloat32(data []byte) ([]float32, error) {
	if len(data)%4 != 0 {
		return nil, fmt.Errorf("invalid data length: must be a multiple of 4")
	}
	count := len(data) / 4
	vector := make([]float32, count)
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.LittleEndian, &vector)
	if err != nil {
		return nil, err
	}
	return vector, nil
}
