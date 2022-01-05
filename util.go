package memfile

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	mrand "math/rand"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/crypto/bcrypt"
)

// The following are several benchmarked implementations of random string
// generators. The information comes directly from this SO answer:
// https://stackoverflow.com/a/31832326
//
// The benchmark results shown in the SO post are:
// BenchmarkRunes-4                     2000000    723 ns/op   96 B/op   2 allocs/op
// BenchmarkBytes-4                     3000000    550 ns/op   32 B/op   2 allocs/op
// BenchmarkBytesRmndr-4                3000000    438 ns/op   32 B/op   2 allocs/op
// BenchmarkBytesMask-4                 3000000    534 ns/op   32 B/op   2 allocs/op
// BenchmarkBytesMaskImpr-4            10000000    176 ns/op   32 B/op   2 allocs/op
// BenchmarkBytesMaskImprSrc-4         10000000    139 ns/op   32 B/op   2 allocs/op
// BenchmarkBytesMaskImprSrcSB-4       10000000    134 ns/op   16 B/op   1 allocs/op
// BenchmarkBytesMaskImprSrcUnsafe-4   10000000    115 ns/op   16 B/op   1 allocs/op
//
// Just by switching from runes to bytes, we immediately have 24% performance gain,
// and memory requirement drops to one third.
//
// Getting rid of rand.Intn() and using rand.Int63() instead gives another 20% boost.
//
// Masking (and repeating in case of big indices) slows down a little (due to
// repetition calls): -22%...
//
// But when we make use of all (or most) of the 63 random bits (10 indices
// from one rand.Int63() call): that speeds up big time: 3 times.
//
// If we settle with a (non-default, new) rand.Source instead of rand.Rand,
// we again gain 21%.
//
// If we utilize strings.Builder, we gain a tiny 3.5% in speed, but we also
// achieved 50% reduction in memory usage and allocations! That's nice!
//
// Finally if we dare to use package unsafe instead of strings.Builder,
// we again gain a nice 14%.
//
// Comparing the final to the initial solution: RandStringBytesMaskImprSrcUnsafe()
// is 6.3 times faster than RandStringRunes(), uses one sixth memory and half as
// few allocations. Mission accomplished.

var (
	letters     = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	mrand.Seed(time.Now().UnixNano())
}

var randUniUri = NewLen

var lettersVar = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// randSeq is the basic random sequence generator from the
// original SO answer (called "Pauls's solution"):
// https://stackoverflow.com/a/22892986
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = lettersVar[mrand.Intn(len(lettersVar))]
	}
	return string(b)
}

// randomString is from a later answer in the SO post.
func randomString(length int) string {
	mrand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

// randomBase64String is from a later answer in the SO post.
func randomBase64String(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/float64(1.33333333333))))
	rand.Read(buff)
	str := base64.RawURLEncoding.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

// randomBase16String is from a later answer in the SO post.
func randomBase16String(l int) string {
	buff := make([]byte, int(math.Ceil(float64(l)/2)))
	rand.Read(buff)
	str := hex.EncodeToString(buff)
	return str[:l] // strip 1 extra character we get from odd length results
}

// RandStr is from a later answer in the SO post.
func RandStr(n int) (str string) {
	b := make([]byte, n)
	rand.Read(b)
	str = fmt.Sprintf("%x", b)
	return
}

var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"

func shortID(length int) string {
	ll := len(chars)
	b := make([]byte, length)
	rand.Read(b) // generates len(b) random bytes
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%ll]
	}
	return string(b)
}

func randStr(len int) string {
	buff := make([]byte, len)
	rand.Read(buff)
	str := base64.StdEncoding.EncodeToString(buff)
	// Base 64 can be longer than len
	return str[:len]
}

var alphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generate(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	for i := 0; i < size; i++ {
		b[i] = alphabet[b[i]/5]
	}
	return *(*string)(unsafe.Pointer(&b))
}

// len(encodeURL) == 64. This allows (x <= 265) x % 64 to have an even
// distribution.
const encodeURL = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// A helper function create and fill a slice of length n with characters from
// a-zA-Z0-9_-. It panics if there are any problems getting random bytes.
func RandAsciiBytes(n int) []byte {
	output := make([]byte, n)

	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)

	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}

	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])

		// random % 64
		randomPos := random % uint8(len(encodeURL))

		// put into output
		output[pos] = encodeURL[randomPos]
	}

	return output
}

func GenerateCertificateNumber() string {
	CertificateLength := 7
	t := time.Now().String()
	CertificateHash, err := bcrypt.GenerateFromPassword([]byte(t), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println(err)
	}
	// Make a Regex we only want letters and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	processedString := reg.ReplaceAllString(string(CertificateHash), "")
	fmt.Println(string(processedString))

	CertificateNumber := strings.ToUpper(string(processedString[len(processedString)-CertificateLength:]))
	fmt.Println(CertificateNumber)
	return CertificateNumber
}

// Doesn't share the rand library globally, reducing lock contention
type Rand struct {
	Seed int64
	Pool *sync.Pool
}

var (
	MRand    = NewRand()
	randlist = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
)

// init random number generator
func NewRand() *Rand {
	p := &sync.Pool{New: func() interface{} {
		return mrand.New(mrand.NewSource(getSeed()))
	},
	}
	mrand := &Rand{
		Pool: p,
	}
	return mrand
}

// get the seed
func getSeed() int64 {
	return time.Now().UnixNano()
}

func (s *Rand) getrand() *mrand.Rand {
	return s.Pool.Get().(*mrand.Rand)
}
func (s *Rand) putrand(r *mrand.Rand) {
	s.Pool.Put(r)
}

// get a random number
func (s *Rand) Intn(n int) int {
	r := s.getrand()
	defer s.putrand(r)

	return r.Intn(n)
}

//  bulk get random numbers
func (s *Rand) Read(p []byte) (int, error) {
	r := s.getrand()
	defer s.putrand(r)

	return r.Read(p)
}

func CreateRandomString(len int) string {
	b := make([]byte, len)
	_, err := MRand.Read(b)
	if err != nil {
		return ""
	}
	for i := 0; i < len; i++ {
		b[i] = randlist[b[i]%(62)]
	}
	return *(*string)(unsafe.Pointer(&b))
}

const (
	chars2   = "0123456789_abcdefghijkl-mnopqrstuvwxyz" //ABCDEFGHIJKLMNOPQRSTUVWXYZ
	charsLen = len(chars2)
	mask     = 1<<6 - 1
)

var rng = mrand.NewSource(time.Now().UnixNano())

// RandStr 返回指定长度的随机字符串
func RandStr2(ln int) string {
	/* chars 38个字符
	 * rng.Int63() 每次产出64bit的随机数,每次我们使用6bit(2^6=64) 可以使用10次
	 */
	buf := make([]byte, ln)
	for idx, cache, remain := ln-1, rng.Int63(), 10; idx >= 0; {
		if remain == 0 {
			cache, remain = rng.Int63(), 10
		}
		buf[idx] = chars2[int(cache&mask)%charsLen]
		cache >>= 6
		remain--
		idx--
	}
	return *(*string)(unsafe.Pointer(&buf))
}

const (
	// 52 possibilities
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// 6 bits to represent 64 possibilities / indexes
	letterIdxBits = 6
	// All 1-bits, as many as letterIdxBits
	letterIdxMask = 1<<letterIdxBits - 1
)

func SecureRandomAlphaString(length int) string {

	result := make([]byte, length)
	bufferSize := int(float64(length) * 1.3)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			randomBytes = SecureRandomBytes(bufferSize)
		}
		if idx := int(randomBytes[j%length] & letterIdxMask); idx < len(letterBytes) {
			result[i] = letterBytes[idx]
			i++
		}
	}

	return string(result)
}

// SecureRandomBytes returns the requested number of bytes using crypto/rand
func SecureRandomBytes(length int) []byte {
	var randomBytes = make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Fatal("Unable to generate random bytes")
	}
	return randomBytes
}

// SecureRandomString returns a string of the requested length,
// made from the byte characters provided (only ASCII allowed).
// Uses crypto/rand for security. Will panic if len(availableCharBytes) > 256.
func SecureRandomString(availableCharBytes string, length int) string {

	// Compute bitMask
	availableCharLength := len(availableCharBytes)
	if availableCharLength == 0 || availableCharLength > 256 {
		panic("availableCharBytes length must be greater than 0 and less than or equal to 256")
	}
	var bitLength byte
	var bitMask byte
	for bits := availableCharLength - 1; bits != 0; {
		bits = bits >> 1
		bitLength++
	}
	bitMask = 1<<bitLength - 1

	// Compute bufferSize
	bufferSize := length + length/3

	// Create random string
	result := make([]byte, length)
	for i, j, randomBytes := 0, 0, []byte{}; i < length; j++ {
		if j%bufferSize == 0 {
			// Random byte buffer is empty, get a new one
			randomBytes = SecureRandomBytes(bufferSize)
		}
		// Mask bytes to get an index into the character slice
		if idx := int(randomBytes[j%length] & bitMask); idx < availableCharLength {
			result[i] = availableCharBytes[idx]
			i++
		}
	}

	return string(result)
}

func RandStringRunes3(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[mrand.Intn(len(letterRunes))]
	}
	return string(b)
}
