package ssss

import (
	"testing"
)

func Test_toInt64(t *testing.T) {
	if toInt64(int8(8)) != int64(8) {
		t.Fatal("type cast, int8 -> int64, fail!")
	}
	if toInt64(int16(16)) != int64(16) {
		t.Fatal("type cast, int16 -> int64, fail!")
	}
	if toInt64(int32(32)) != int64(32) {
		t.Fatal("type cast, int32 -> int64, fail!")
	}
	if toInt64(int64(64)) != int64(64) {
		t.Fatal("type cast, int64 -> int64, fail!")
	}

	if toInt64(uint8(8)) != int64(8) {
		t.Fatal("type cast, uint8 -> int64, fail!")
	}
	if toInt64(uint16(16)) != int64(16) {
		t.Fatal("type cast, uint16 -> int64, fail!")
	}
	if toInt64(uint32(32)) != int64(32) {
		t.Fatal("type cast, uint32 -> int64, fail!")
	}
	if toInt64(uint64(64)) != int64(64) {
		t.Fatal("type cast, uint64 -> int64, fail!")
	}

	if toInt64(float32(32.99)) != int64(32) {
		t.Fatal("type cast, float32 -> int64, fail!")
	}
	if toInt64(float64(64.99)) != int64(64) {
		t.Fatal("type cast, float64 -> int64, fail!")
	}

	if toInt64(true) != int64(1) {
		t.Fatal("type cast, bool -> int64, fail!")
	}
	if toInt64(false) != int64(0) {
		t.Fatal("type cast, bool -> int64, fail!")
	}

	if toInt64("231") != int64(231) {
		t.Fatal("type cast, string -> int64, fail!")
	}
}

func Test_toUint64(t *testing.T) {
	if toUint64(int8(8)) != uint64(8) {
		t.Fatal("type cast, int8 -> uint64, fail!")
	}
	if toUint64(int16(16)) != uint64(16) {
		t.Fatal("type cast, int16 -> uint64, fail!")
	}
	if toUint64(int32(32)) != uint64(32) {
		t.Fatal("type cast, int32 -> uint64, fail!")
	}
	if toUint64(int64(64)) != uint64(64) {
		t.Fatal("type cast, int64 -> uint64, fail!")
	}

	if toUint64(uint8(8)) != uint64(8) {
		t.Fatal("type cast, uint8 -> uint64, fail!")
	}
	if toUint64(uint16(16)) != uint64(16) {
		t.Fatal("type cast, uint16 -> uint64, fail!")
	}
	if toUint64(uint32(32)) != uint64(32) {
		t.Fatal("type cast, uint32 -> uint64, fail!")
	}
	if toUint64(uint64(64)) != uint64(64) {
		t.Fatal("type cast, uint64 -> uint64, fail!")
	}

	if toUint64(float32(32.99)) != uint64(32) {
		t.Fatal("type cast, float32 -> uint64, fail!")
	}
	if toUint64(float64(64.99)) != uint64(64) {
		t.Fatal("type cast, float64 -> uint64, fail!")
	}

	if toUint64(true) != uint64(1) {
		t.Fatal("type cast, bool -> uint64, fail!")
	}
	if toUint64(false) != uint64(0) {
		t.Fatal("type cast, bool -> uint64, fail!")
	}

	if toUint64("231") != uint64(231) {
		t.Fatal("type cast, string -> uint64, fail!")
	}
}

func Test_toFloat64(t *testing.T) {
	if toFloat64(int8(8)) != float64(8) {
		t.Fatal("type cast, int8 -> float64, fail!")
	}
	if toFloat64(int16(16)) != float64(16) {
		t.Fatal("type cast, int16 -> float64, fail!")
	}
	if toFloat64(int32(32)) != float64(32) {
		t.Fatal("type cast, int32 -> float64, fail!")
	}
	if toFloat64(int64(64)) != float64(64) {
		t.Fatal("type cast, int64 -> float64, fail!")
	}

	if toFloat64(uint8(8)) != float64(8) {
		t.Fatal("type cast, uint8 -> float64, fail!")
	}
	if toFloat64(uint16(16)) != float64(16) {
		t.Fatal("type cast, uint16 -> float64, fail!")
	}
	if toFloat64(uint32(32)) != float64(32) {
		t.Fatal("type cast, uint32 -> float64, fail!")
	}
	if toFloat64(uint64(64)) != float64(64) {
		t.Fatal("type cast, uint64 -> float64, fail!")
	}

	if toFloat64(float32(32.95)) != float64(32.95000076293945) {
		t.Fatal("type cast, float32 -> float64, fail!", toFloat64(float32(32.95)), float64(32.95))
	}
	if toFloat64(float64(64.99)) != float64(64.99) {
		t.Fatal("type cast, float64 -> float64, fail!")
	}

	if toFloat64(true) != float64(1) {
		t.Fatal("type cast, bool -> float64, fail!")
	}
	if toFloat64(false) != float64(0) {
		t.Fatal("type cast, bool -> float64, fail!")
	}

	if toFloat64("231.66") != float64(231.66) {
		t.Fatal("type cast, string -> float64, fail!", toFloat64("231.66"))
	}
}
