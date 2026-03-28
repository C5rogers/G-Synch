package tests

import (
	"testing"

	"github.com/C5rogers/G-Synch/internal/audit/adapters/pg"
	"github.com/stretchr/testify/assert"
)

func TestSerializePrimaryKey_SingleColumn(t *testing.T) {
	row := []interface{}{1, "john@example.com", "John"}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 0, "email": 1, "name": 2}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "1", result)
}

func TestSerializePrimaryKey_CompositeKey(t *testing.T) {
	row := []interface{}{10, 20, "data"}
	pkCols := []string{"user_id", "role_id"}
	colIndex := map[string]int{"user_id": 0, "role_id": 1, "info": 2}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "10::20", result)
}

func TestSerializePrimaryKey_MissingColumn(t *testing.T) {
	row := []interface{}{1, "value"}
	pkCols := []string{"nonexistent"}
	colIndex := map[string]int{"id": 0, "value": 1}

	_, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestSerializePrimaryKey_IndexOutOfRange(t *testing.T) {
	row := []interface{}{1}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 5}

	_, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.Error(t, err)
}

func TestSerializePrimaryKey_StringPK(t *testing.T) {
	row := []interface{}{"uuid-123-abc", "data"}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 0, "data": 1}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "uuid-123-abc", result)
}

func TestSerializePrimaryKey_NilValue(t *testing.T) {
	row := []interface{}{nil, "data"}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 0, "data": 1}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "<nil>", result)
}

func TestSerializePrimaryKey_UUIDBytes(t *testing.T) {
	row := []interface{}{[]byte{118, 88, 77, 15, 84, 235, 67, 211, 187, 100, 140, 155, 227, 6, 185, 238}}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 0}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "76584d0f-54eb-43d3-bb64-8c9be306b9ee", result)
}

func TestSerializePrimaryKey_ByteSliceString(t *testing.T) {
	row := []interface{}{[]byte("team-42")}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 0}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "team-42", result)
}

func TestSerializePrimaryKey_UUIDArray(t *testing.T) {
	row := []interface{}{[16]uint8{118, 88, 77, 15, 84, 235, 67, 211, 187, 100, 140, 155, 227, 6, 185, 238}}
	pkCols := []string{"id"}
	colIndex := map[string]int{"id": 0}

	result, err := pg.SerializePrimaryKey(row, pkCols, colIndex)
	assert.NoError(t, err)
	assert.Equal(t, "76584d0f-54eb-43d3-bb64-8c9be306b9ee", result)
}

func TestAdapterEngine(t *testing.T) {
	adapter := &pg.Adapter{}
	assert.Equal(t, "postgres", adapter.Engine())
}
