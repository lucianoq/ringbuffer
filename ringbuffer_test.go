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

		inputBuffer *RingBuffer
		toWrite     []byte

		wantBuffer *RingBuffer
		wantErr    bool
	}{
		{
			name:        "write nil",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     nil,
			wantBuffer: &RingBuffer{
				buf:      []byte{0, 0, 0},
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write empty slice",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte{},
			wantBuffer: &RingBuffer{
				buf:      []byte{0, 0, 0},
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write less than size",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte("a"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 0, 0},
				pos:      1,
				written:  1,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write to fill length, no grow",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte("abc"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c'},
				pos:      3,
				written:  3,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write more than size, grow double",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte("abcde"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 0},
				pos:      5,
				written:  5,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write more than size, grow double, starting size 0",
			inputBuffer: NewRingBuffer(0, 7),
			toWrite:     []byte("a"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a'},
				pos:      1,
				written:  1,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write more than size, grow double many times, starting size 0",
			inputBuffer: NewRingBuffer(0, 7),
			toWrite:     []byte("abcde"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 0, 0},
				pos:      5,
				written:  5,
				ringMode: false,
				maxSize:  7,
			},
		},
		{
			name:        "write to fill cap, grow at max",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte("abcdefg"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'},
				pos:      0,
				written:  7,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name:        "write to exceed cap, grow at max",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte("abcdefghijk"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'e', 'f', 'g', 'h', 'i', 'j', 'k'},
				pos:      0,
				written:  11,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name: "write less the max in a ring mode, starting from middle",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'},
				pos:      2,
				written:  14,
				ringMode: true,
				maxSize:  7,
			},
			toWrite: []byte("123"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', '1', '2', '3', 'f', 'g'},
				pos:      5,
				written:  17,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name: "write to fill in a ring mode, starting from middle",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'},
				pos:      2,
				written:  14,
				ringMode: true,
				maxSize:  7,
			},
			toWrite: []byte("12345"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', '1', '2', '3', '4', '5'},
				pos:      7,
				written:  19,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name: "write to overflow in a ring mode, starting from middle",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'},
				pos:      2,
				written:  14,
				ringMode: true,
				maxSize:  7,
			},
			toWrite: []byte("123456"),
			wantBuffer: &RingBuffer{
				buf:      []byte{'6', 'b', '1', '2', '3', '4', '5'},
				pos:      1,
				written:  20,
				ringMode: true,
				maxSize:  7,
			},
		},
		{
			name:        "write to exceed cap many times",
			inputBuffer: NewRingBuffer(3, 7),
			toWrite:     []byte("abcdefghijklmnopqrstuvwxyz"),
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

			gotN, err := tt.inputBuffer.Write(tt.toWrite)

			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotN != len(tt.toWrite) {
				t.Errorf("Written %d bytes, expected %d", gotN, len(tt.toWrite))
				return
			}

			if !reflect.DeepEqual(tt.inputBuffer, tt.wantBuffer) {
				t.Errorf("Write() got = %+v want %+v", tt.inputBuffer, tt.wantBuffer)
			}
		})
	}
}

func TestRingBuffer_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputBuffer *RingBuffer
		want        string
	}{
		{
			name: "empty",
			inputBuffer: &RingBuffer{
				buf:      []byte{},
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  4,
			},
			want: "",
		},
		{
			name: "shorter than buf",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 0},
				pos:      3,
				written:  3,
				ringMode: false,
				maxSize:  4,
			},
			want: "abc",
		},
		{
			name: "full buff, no ring",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd'},
				pos:      4,
				written:  4,
				ringMode: false,
				maxSize:  4,
			},
			want: "abcd",
		},
		{
			name: "full buff, ring mode, start on end",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd'},
				pos:      4,
				written:  4,
				ringMode: true,
				maxSize:  4,
			},
			want: "abcd",
		},
		{
			name: "full buff, ring mode, start on 0",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', 'c', 'd'},
				pos:      0,
				written:  4,
				ringMode: true,
				maxSize:  4,
			},
			want: "abcd",
		},
		{
			name: "full buf, ring mode, start in the middle",
			inputBuffer: &RingBuffer{
				buf:      []byte{'e', 'b', 'c', 'd'},
				pos:      1,
				written:  5,
				ringMode: true,
				maxSize:  4,
			},
			want: "bcde",
		},
		{
			name: "full buf, ring mode, start in the middle",
			inputBuffer: &RingBuffer{
				buf:      []byte{'a', 'b', '1', '2', '3', 'f', 'g'},
				pos:      5,
				written:  17,
				ringMode: true,
				maxSize:  7,
			},
			want: "fgab123",
		},
	}
	for _, tt := range tests {
		var tt = tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.inputBuffer.String()

			if got != tt.want {
				t.Errorf("Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRingBuffer_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputBuffer *RingBuffer
		wantBuffer  *RingBuffer
		wantErr     bool
	}{
		{
			name: "ok",
			inputBuffer: &RingBuffer{
				buf:      []byte{'e', 'b', 'c', 'd'},
				pos:      1,
				written:  5,
				ringMode: true,
				maxSize:  4,
			},
			wantBuffer: &RingBuffer{
				buf:      nil,
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		var tt = tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotErr := tt.inputBuffer.Close()

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", gotErr, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.inputBuffer, tt.wantBuffer) {
				t.Errorf("Write() got = %+v want %+v", tt.inputBuffer, tt.wantBuffer)
			}
		})
	}
}

func TestRingBuffer_Reset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		inputBuffer *RingBuffer
		wantBuffer  *RingBuffer
	}{
		{
			name: "ok",
			inputBuffer: &RingBuffer{
				buf:      []byte{'e', 'b', 'c', 'd'},
				pos:      1,
				written:  5,
				ringMode: true,
				maxSize:  4,
			},
			wantBuffer: &RingBuffer{
				buf:      []byte{'e', 'b', 'c', 'd'},
				pos:      0,
				written:  0,
				ringMode: false,
				maxSize:  4,
			},
		},
	}
	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.inputBuffer.Reset()

			if !reflect.DeepEqual(tt.inputBuffer, tt.wantBuffer) {
				t.Errorf("Write() got = %+v want %+v", tt.inputBuffer, tt.wantBuffer)
			}
		})
	}
}
