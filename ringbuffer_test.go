package ringbuffer

import (
	"reflect"
	"testing"
)

func TestNewRingBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		initialSize int
		maxSize     int

		want *RingBuffer
	}{
		{
			name:        "ok",
			initialSize: 10,
			maxSize:     20,
			want: &RingBuffer{
				buf:      make([]byte, 10),
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  20,
			},
		},
		{
			name:        "len greater than cap",
			initialSize: 20,
			maxSize:     10,
			want: &RingBuffer{
				buf:      make([]byte, 10),
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  10,
			},
		},
	}
	for _, tt := range tests {
		var tt = tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewRingBuffer(tt.initialSize, tt.maxSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRingBuffer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRingBuffer_Write(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string

		initialBuffer *RingBuffer
		toWrite       []byte
		wantBuffer    *RingBuffer
		wantErr       bool
	}{
		{
			name:          "write nil",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       nil,
			wantBuffer: &RingBuffer{
				buf:      []byte{0, 0, 0},
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:          "write empty slice",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte{},
			wantBuffer: &RingBuffer{
				buf:      []byte{0, 0, 0},
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:          "write less than size",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte("a"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 0, 0},
				pos:      1,
				written:  1,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:          "write to fill length, no grow",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte("abc"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c'},
				pos:      3,
				written:  3,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:          "write more than size, grow double",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte("abcde"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 0},
				pos:      5,
				written:  5,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:          "write to fill cap, grow at max",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte("abcdefg"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'},
				pos:      0,
				written:  7,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name:          "write to exceed cap, grow at max",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte("abcdefghijk"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'e', 'f', 'g', 'h', 'i', 'j', 'k'},
				pos:      0,
				written:  11,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name:          "write to exceed cap many times",
			initialBuffer: NewRingBuffer(3, 7),
			toWrite:       []byte("abcdefghijklmnopqrstuvwxyz"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'t', 'u', 'v', 'w', 'x', 'y', 'z'},
				pos:      0,
				written:  26,
				ringMode: true,
				maxSize:  7,
			},
		},
	}
	for _, tt := range tests {
		var tt = tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotN, err := tt.initialBuffer.Write(tt.toWrite)

			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotN != len(tt.toWrite) {
				t.Errorf("Written %d bytes, expected %d", gotN, len(tt.toWrite))
				return
			}

			if !reflect.DeepEqual(tt.initialBuffer, tt.wantBuffer) {
				t.Errorf("Write() got = %+v want %+v", tt.initialBuffer, tt.wantBuffer)
			}
		})
	}
}
