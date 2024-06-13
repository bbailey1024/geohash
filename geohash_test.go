package geohash

import (
	"math"
	"testing"
)

// Benchmark constants
const (
	testBits          = 64
	testHash          = "dngb2x6mnetr"
	testHashHighPrec  = "dngb2x6mnetr3zzycbjt"
	testHashInt       = 0x651ea174d3a37371
	testLat           = 38.05339909138269
	testLng           = -84.70121386485815
	testPrecision     = 12
	testPrecisionHigh = 20
)

type TestCase struct {
	hashInt      uint64
	hash         string
	hashHighPrec string
	lat, lng     float64
}

var testCases = []TestCase{
	{0xc28a4d93b20a22f8, "sb54v4xk18jg", "sb54v4xk18jgjuhjcjpb", 0.497818518, 38.198505253},
	{0x003558b7d15148f1, "00upjeyjb54g", "00upjeyjb54g36jzuutr", -84.529178182, -174.125057287},
	{0x949dcd034ca43b30, "kkfwu0udnhxm", "kkfwu0udnhxm00jegdxy", -17.090238388, 14.947853282},
	{0x7d44be93dc8c3d1f, "gp2cx4ywjhyj", "gp2cx4ywjhyjyzv06u0h", 86.06108453, -43.628546008},
	{0x801e4ccef502f590, "h0g4tmrp0cut", "h0g4tmrp0cut06wq5n34", -85.311745894, 4.459114168},
	{0xd90e166bb45673b6, "v471duxnbttv", "v471duxnbttvewdns8h9", 57.945830289, 49.349241965},
	{0x81d1418fe43c871f, "h78n33z47k3j", "h78n33z47k3jzqj9jbpx", -69.203844118, 11.314685805},
	{0x7ef3c3f9b722a109, "gvtw7yer4bhh", "gvtw7yer4bhhky4g2vtw", 77.073040753, -3.346243298},
	{0x03adcf02afef7916, "0fqwy0pgxxwj", "0fqwy0pgxxwjdcff38h8", -76.156584583, -136.834730089},
	{0x644a3e6ab5fbafd6, "dj53wuppzfrx", "dj53wuppzfrxeu3esdgw", 28.411988257, -85.123100792},
	{0xbcd540239bd297fb, "rmbn08wvubcz", "rmbn08wvubczqgew688g", -11.597823607, 146.281448853},
	{0xb663a5ee6c8fa0df, "qtjucvmdjyhe", "qtjucvmdjyhezt84b2y1", -16.010823784, 120.67064801},
	{0x4f24b9f758e2851e, "9wkcmxuswb2j", "9wkcmxuswb2jxvb5x9bq", 35.419323354, -105.572143468},
	{0x49c85bd32bc93b07, "9745rntct4xh", "9745rntct4xhfr485q4v", 17.482266365, -120.621762327},
	{0x73213a6da3c7ca39, "fdhmnve3sz53", "fdhmnve3sz53knyssgg0", 57.159413941, -61.222135062},
	{0x0f044c9a65ca1bc2, "1w24t6m5t8ew", "1w24t6m5t8ew5n2zxyrw", -54.391332719, -112.262179799},
	{0x575f88776cb53f16, "bxgshxvdqnzj", "bxgshxvdqnzjdj6wwey0", 89.33987042, -152.372551026},
	{0xf435c4f4150cf21b, "yhuw9x0p1mt1", "yhuw9x0p1mt1rjyjzfs5", 72.901011648, 96.39410362},
}

func TestEncodeStrConcat(t *testing.T) {
	for _, c := range testCases {
		res, err := encodeStrConcat(c.lat, c.lng, testPrecision)
		if err != nil {
			t.Errorf("%s", err.Error())
		}

		if res != c.hash {
			t.Errorf("Encode = %s, want %s", res, c.hash)
		}
	}
}

func TestEncodeDoubling(t *testing.T) {
	for _, c := range testCases {
		res := encodeDoubling(c.lat, c.lng, testPrecision)

		if res != c.hash {
			t.Errorf("Encode = %s, want %s", res, c.hash)
		}
	}
}

func TestEncode(t *testing.T) {
	for _, c := range testCases {
		res := Encode(c.lat, c.lng)

		if res != c.hash {
			t.Errorf("Encode = %s, want %s", res, c.hash)
		}
	}
}

func TestEncodePrecision(t *testing.T) {
	for _, c := range testCases {
		res := EncodePrecision(c.lat, c.lng, testPrecision)

		if res != c.hash {
			t.Errorf("Encode = %s, want %s", res, c.hash)
		}
	}
}

func TestEncodeHighPrecision(t *testing.T) {
	for _, c := range testCases {
		res := EncodeHighPrecision(c.lat, c.lng, testPrecisionHigh)

		if res != c.hashHighPrec {
			t.Errorf("Encode = %s, want %s", res, c.hashHighPrec)
		}
	}
}

func TestEncodeInt(t *testing.T) {
	for _, c := range testCases {
		res := EncodeInt(c.lat, c.lng)

		if res != c.hashInt {
			t.Errorf("Encode = %x, want %x", res, c.hashInt)
		}
	}
}

func TestEncodeIntPrecision(t *testing.T) {
	for _, c := range testCases {
		res := EncodeIntPrecision(c.lat, c.lng, testBits)

		if res != c.hashInt {
			t.Errorf("Encode = %x, want %x", res, c.hashInt)
		}
	}
}

func TestDecode(t *testing.T) {
	for _, c := range testCases {
		lat, lng := Decode(c.hash)

		f := 0.000001

		if math.Abs(lat-c.lat) > f || math.Abs(lng-c.lng) > f {
			t.Errorf("Decode = %.6f, %.6f, want %.6f, %.6f", lat, lng, c.lat, c.lng)
		}
	}
}

func TestDecodeHighPrecision(t *testing.T) {
	for _, c := range testCases {
		lat, lng := DecodeHighPrecision(c.hashHighPrec)

		f := 0.000000001

		if math.Abs(lat-c.lat) > f || math.Abs(lng-c.lng) > f {
			t.Errorf("Decode = %.9f, %.9f, want %.9f, %.9f", lat, lng, c.lat, c.lng)
		}
	}
}

func TestDecodeInt(t *testing.T) {
	for _, c := range testCases {
		lat, lng := DecodeInt(c.hashInt)

		f := 0.000001

		if math.Abs(lat-c.lat) > f || math.Abs(lng-c.lng) > f {
			t.Errorf("Decode = %.6f, %.6f, want %.6f, %.6f", lat, lng, c.lat, c.lng)
		}
	}
}

func BenchmarkEncodeStrConcat(b *testing.B) {
	for n := 0; n < b.N; n++ {
		encodeStrConcat(testLat, testLng, testPrecision)
	}
}

func BenchmarkEncodeDoubling(b *testing.B) {
	for n := 0; n < b.N; n++ {
		encodeDoubling(testLat, testLng, testPrecision)
	}
}

func BenchmarkEncodeBitwiseOr(b *testing.B) {
	for n := 0; n < b.N; n++ {
		encodeBitwiseOr(testLat, testLng, testPrecision)
	}
}

func BenchmarkEncode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Encode(testLat, testLng)
	}
}

func BenchmarkEncodePrecision(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodePrecision(testLat, testLng, testPrecision)
	}
}

func BenchmarkEncodeHighPrecision(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodeHighPrecision(testLat, testLng, testPrecisionHigh)
	}
}

func BenchmarkEncodeInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodeInt(testLat, testLng)
	}
}

func BenchmarkEncodeIntPrecision(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodeIntPrecision(testLat, testLng, testBits)
	}
}

func BenchmarkDecode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Decode(testHash)
	}
}

func BenchmarkDecodeHighPrecision(b *testing.B) {
	for n := 0; n < b.N; n++ {
		DecodeHighPrecision(testHashHighPrec)
	}
}

func BenchmarkDecodeInt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		DecodeInt(testHashInt)
	}
}
