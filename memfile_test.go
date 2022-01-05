package memfile

import (
	"io"
	mrand "math/rand"
	"sync"
	"testing"
	"time"
)

const (
	testFileName = "testfile.test.txt"
)

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[mrand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.Intn(len(letterBytes))]
	}
	return string(b)
}

// This works and is significantly faster, the disadvantage is that the probability of all the letters will not be exactly the same (assuming rand.Int63() produces all 63-bit numbers with equal probability). Although the distortion is extremely small as the number of letters 52 is much-much smaller than 1<<63 - 1, so in practice this is perfectly fine.
func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func init() {
	mrand.Seed(time.Now().UnixNano())

	// b := []byte{}
	// for i in rand.Intn() {
	// 	b = append(b,byte(i))
	// }
	// io.WriteFile(testfile)
}

// memFileSamples := []struct {
// 	NewMemFile("testfile.test.txt")
// }

func Test_memFile_Close(t *testing.T) {
	type fields struct {
		filename        string
		ReadWriteCloser io.ReadWriteCloser
		mu              sync.RWMutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
		{"fake", fields{"fake", nil, sync.RWMutex{}}, false},
		{"", fields{}, false},
		{"", fields{}, false},
		{"", fields{}, false},
		{"", fields{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &memFile{
				filename:        tt.fields.filename,
				ReadWriteCloser: tt.fields.ReadWriteCloser,
				mu:              tt.fields.mu,
			}
			if err := m.Close(); (err != nil) != tt.wantErr {
				t.Errorf("memFile.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
