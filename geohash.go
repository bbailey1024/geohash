// Package geohash implements the geohashing algorithm to produce geohash strings and integers.
// While this library may be used for internal projects, it should be considered academic.
// This package explores and documents various methods related to geohashing.
// See https://github.com/bbailey1024/geohash/blob/master/README.md for details and references.
// If a production geohashing library is required, consider https://github.com/mmcloughlin/geohash.
package geohash

import (
	"bytes"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	base32        = "0123456789bcdefghjkmnpqrstuvwxyz"
	bitsMin       = 1
	bitsMax       = 64
	latMax        = 90.0
	lngMax        = 180.0
	precisionMin  = 1
	precisionMax  = 12
	precisionHigh = 20
)

// bit positions for a 5-bit geohash character utilized by encodeBitwiseOr
var bitPositions = []int{16, 8, 4, 2, 1}

// Encode returns a geohash string of the lat, lng coordinates based on the max character precision of 12.
func Encode(lat, lng float64) string {
	return encode(lat, lng, precisionMax)
}

// EncodePrecision returns a geohash string of the lat, lng coordinates based on the provided character precision.
// Acceptable precision values are 1 to 12 characters.
func EncodePrecision(lat, lng float64, precision int) string {
	precision = validate(precisionMin, precisionMax, precision)
	return encode(lat, lng, precision)
}

// EncodeHighPrecision returns a geohash string of the lat, lng coordinates based on the provided character precision.
// Leverages a slightly slower method for generating hashes, but allows precision up to 20 characters.
func EncodeHighPrecision(lat, lng float64, precision int) string {
	precision = validate(precisionMin, precisionHigh, precision)
	return encodeBitwiseOr(lat, lng, precision)
}

// EncodeInt returns a uint64 geohash of lat, lng coordinates based on the max bit precision of 64.
func EncodeInt(lat, lng float64) uint64 {
	return encodeInt(lat, lng, bitsMax)
}

// EncodeIntPrecision returns a uint64 geohash of lat, lng coordinates based on the provided bit precision.
// Acceptable bit values are 1 to 64.
// Consider using a multiple of 5 if this value will be encoded as a string using EncodeIntToStr.
func EncodeIntPrecision(lat, lng float64, bits int) uint64 {
	bits = validate(bitsMin, bitsMax, bits)
	return encodeInt(lat, lng, bits)
}

// EncodeIntToStr converts an integer hash to a geohash string with precision characters.
// This assumes the integer was generated using precision*5 bits.
// If this is not the case, the resulting geohash string will be malformed.
// A 64-bit precision integer should be right shifted 4 to get a 12 character hash.
func EncodeIntToStr(hash uint64, precision int) string {
	return encodeIntToStr(hash)[precisionMax-precision:]
}

// EncodeStrToInt converts a geohash string of any precision to a geohash integer.
func EncodeStrToInt(hash string) uint64 {
	return encodeStrToInt(hash)
}

// Decode returns the estimated lat, lng coordinates of a geohash string up to a precision of 12 characters.
// Exceeding character limit will truncate the geohash string to the precision max of 12 characters.
func Decode(hash string) (float64, float64) {
	if len(hash) > precisionMax {
		hash = hash[:precisionMax]
	}
	return decode(hash)
}

// DecodeHighPrecision returns the estimated lat, lng coordinates for a geohash string up to a precision of 20 chartacters.
// Exceeding character limit will truncate the geohash string to the precision max of 20 characters.
// Uses a slightly slower decoding algorithm than Decode.
func DecodeHighPrecision(hash string) (float64, float64) {
	if len(hash) > precisionHigh {
		hash = hash[:precisionHigh]
	}
	return decodeBits(hash)
}

// DecodeInt returns the estimated lat, lng coordinates for a geohash integer.
// Assumes max precision of 64 bits.
func DecodeInt(hash uint64) (float64, float64) {
	return decodeInt(hash, bitsMax)
}

// DecodeIntPrecision returns the estimated lat, lng coordinates for a geohash integer of specified precision.
func DecodeIntPrecision(hash uint64, bits int) (float64, float64) {
	return decodeInt(hash, bits)
}

// decode returns the estimated lat, lng coordinates for a geohash string of any precision.
// The length of the hash is used to derive the bit precision for decodeInt.
func decode(hash string) (float64, float64) {
	hashInt := encodeStrToInt(hash)
	return decodeInt(hashInt, len(hash)*5)
}

// decodeInt returns the estimated lat, lng coordinates by deinterleaving the uint64 to their respective uint32 values.
// The uint32 values are decoded using decodeRange to return the latitude and longitude values.
func decodeInt(hash uint64, bits int) (float64, float64) {
	lat32, lng32 := deinterleave(hash << (64 - bits))
	lat := decodeRange(lat32, latMax)
	lng := decodeRange(lng32, lngMax)
	return lat, lng
}

// decodeRange denormalizes x (uint32 of lat or lng) based on its range (90 or 180).
func decodeRange(x uint32, r float64) float64 {
	p := float64(x) / math.Exp2(32)
	return 2*r*p - r
}

// decodeBits returns the estimated lat, lng coordinates for a geohash string of any precision.
// Each bit of every 5-bit character of the geohash string is evaluated.
// Starting with bit 5, each character is shifted right by 4, decrementing to 0.
// This moves each bit to the zero position in sequence.
// A bitwise and operation is performed using this value and 1 to determine if the bit is 0 or 1.
// Each iteration produces a box of min/max values for lat/lng respectively.
// The center of this box is returned as the estimated point of the geohash string.
func decodeBits(hash string) (float64, float64) {
	latmin, latmax := -latMax, latMax
	lngmin, lngmax := -lngMax, lngMax
	even := true

	for i := range hash {
		idx := bytes.IndexByte([]byte(base32), hash[i])

		for j := 4; j > -1; j-- {
			bit := idx >> j & 1

			if even {
				mid := (lngmin + lngmax) / 2
				if bit == 1 {
					lngmin = mid
				} else {
					lngmax = mid
				}
			} else {
				mid := (latmin + latmax) / 2
				if bit == 1 {
					latmin = mid
				} else {
					latmax = mid
				}
			}
			even = !even
		}
	}

	// Could return a bounding box here using sw: min values, ne: max values

	lat := float64((latmin + latmax) / 2)
	lng := float64((lngmin + lngmax) / 2)
	return lat, lng
}

// encode returns a geohash string of desired character precision based on provided lat, lng coordinates.
// Bit precision for encodeInt is based on the 5-bit geohash character.
// The string returned by encodeIntToStr will be 0 left padded based on requested precision.
// Character precision is achieved by discarding these values by adjusting slice start.
// Example: precison = 7, hash = 00000c7hxfhb, start is 12-7, discarding the leading 5 zero values.
func encode(lat, lng float64, precision int) string {
	bits := precision * 5
	hash := encodeInt(lat, lng, bits)
	return encodeIntToStr(hash)[precisionMax-precision:]
}

// encodeInt normalizes lat, lng coordinates as uint32 values then interleaves a uint64 hash.
// Returns a right shifted uint64 to achieve the desired bit precision.
func encodeInt(lat, lng float64, bits int) uint64 {
	lat32 := encodeRange(lat, latMax)
	lng32 := encodeRange(lng, lngMax)
	hash := interleave(lat32, lng32)
	return hash >> (64 - bits)
}

// encodeRange normalizes x (lat or lng) based on its range (90 or 180) into to [0,1] as a uint32.
func encodeRange(x, r float64) uint32 {
	return uint32(math.Floor(math.Exp2(32) * (x + r) / (r * 2)))
}

// encodeIntToStr returns a 12 character geohash string based on the provided uint64.
// On each iteration, a bitwise AND operation is performed on the uint64 hash using 0x1f (11111).
// The resulting 5-bit value is encoded to its respective geohash character and stored in the byte array in reverse order.
// The hash integer is right shifted 5 bits allowing the next iteration to retrieve the next 5-bit character.
// This results in zero left padding for hash integers with lower precision.
func encodeIntToStr(hashInt uint64) string {
	b := [12]byte{}

	for i := 0; i < 12; i++ {
		b[11-i] = base32[hashInt&0x1f]
		hashInt >>= 5
	}

	return string(b[:])
}

// encodeStrToInt returns a geohash uint64 based on the provided geohash string.
// For every character, the hashInt value is left shifted by 5 (each geohash character is 5 bits).
// This moves the binary value of the prior character to the next position to the left.
// A bitwise or of the hashInt and the base32 integer value of the next characer is done, thus creating the new value.
// The next iteration shifts left 5 and gets the next character until all characters are finished.
// The initial left shift of 5 in the loop is done on a zero value integer which is zero.
func encodeStrToInt(hashStr string) uint64 {
	var hashInt uint64
	for _, c := range hashStr {
		hashInt = (hashInt << 5) | uint64(slices.Index([]rune(base32), c))
	}
	return hashInt
}

// interleave generates a uint64 from the uint32 values for lat and lng.
// Reference: https://graphics.stanford.edu/~seander/bithacks.html#InterleaveBMN
func interleave(lat32, lng32 uint32) uint64 {

	// Example
	// lat, lng: 63.263412836, -117.333484316
	// lat32, lng32: 0xd9f98174, 0x2c90174d

	// lat32 binary: 11011001111110011000000101110100
	// lng32 binary: 00101100100100000001011101001101

	// The goal is to create a 64-bit number by alternating each bit from the 32-bit numbers.
	// First step is to spread out each 32-bit value to a 64-bit one by adding a zero before every bit.
	// The added zeros will eventually allow the two 32-bit integers to be merged.
	// Here is what the lat32 value will look like after being spread:
	//	0101000101000001010101010100000101000000000000010001010100010000
	// The issue is that each bit must be moved by a different amount.
	// The first bit must be moved 31 places to the left.
	// The second bit must be moved 29 places to the left and so on.
	// By tracking the desired position of the first bit, the step values can be defined.
	// Bit movement is accomplished by left shifting bits in steps of 16, 8, 2, and 1 (totaling 31).
	// After each shift, a bitwise or operation is performed using the original value.
	// This copies the bits that did not require a shift during this step back to their original position.
	// Afterwards, a bitwise and operation using a bit mask alternating 0 and 1 values by the shift value is applied.
	// This zeros the moved bits that did not require moving based on the current step.

	// Example for the first 16-bit step on lat32 (x).
	// Pipe characters are added at step intervals to aid visualization.
	// Values are 0 padded to make 64 bit comparison simple.
	// x          : 0000000000000000|0000000000000000|1101100111111001|1000000101110100 // Original 32-bit value
	// x << 16    : 0000000000000000|1101100111111001|1000000101110100|0000000000000000 // Left shift 16 adds 16 zeros on right
	// x | x << 16: 0000000000000000|1101100111111001|1101100111111101|1000000101110100 // Or operation copies original bits that didn't require this step
	// 16-bit mask: 0000000000000000|1111111111111111|0000000000000000|1111111111111111 // Alternating 0 and 1 in 16-bit chunks
	// And op     : 0000000000000000|1101100111111001|0000000000000000|1000000101110100 // And operation zeros the bits that did not require this step

	// At the end of the first step, the first 16 bits of the 32-bit integer have been left shifted by 16-bits.
	// A 16-bit gap (zeros) is added, followed by the last 16-bits of the original 16-bit number.
	// The next step is applied using an 8 bit left shift.
	// Visualization aids include a pipe character every 8 bits.

	// x          : 00000000|00000000|11011001|11111001|00000000|00000000|10000001|01110100 // Shifted value from step 1
	// x << 8     : 00000000|11011001|11111001|00000000|00000000|10000001|01110100|00000000 // Left shift adds 8 zeros on right
	// x | x << 8 : 00000000|11011001|11111001|11111001|00000000|10000001|11110101|01110100 // Or operations copies original bits that didn't require this step
	// 8-bit mask : 00000000|11111111|00000000|11111111|00000000|11111111|00000000|11111111 // Alternating 0 and 1 in 8-bit chunks
	// And op     : 00000000|11011001|00000000|11111001|00000000|10000001|00000000|01110100 // And operation zeros the bits that did not require this step

	// x          : 0000|0000|1101|1001|0000|0000|1111|1001|0000|0000|1000|0001|0000|0000|0111|0100
	// x << 4     : 0000|1101|1001|0000|0000|1111|1001|0000|0000|1000|0001|0000|0000|0111|0100|0000
	// x | x << 4 : 0000|1101|1101|1001|0000|1111|1111|1001|0000|1000|1001|0001|0000|0111|0111|0100
	// 4-bit mask : 0000|1111|0000|1111|0000|1111|0000|1111|0000|1111|0000|1111|0000|1111|0000|1111
	// And op     : 0000|1101|0000|1001|0000|1111|0000|1001|0000|1000|0000|0001|0000|0111|0000|0100

	// x          : 00|00|11|01|00|00|10|01|00|00|11|11|00|00|10|01|00|00|10|00|00|00|00|01|00|00|01|11|00|00|01|00
	// x << 2     : 00|11|01|00|00|10|01|00|00|11|11|00|00|10|01|00|00|10|00|00|00|00|01|00|00|01|11|00|00|01|00|00
	// x | x << 2 : 00|11|11|01|00|10|11|01|00|11|11|11|00|10|11|01|00|10|10|00|00|00|01|01|00|01|11|11|00|01|01|00
	// 2-bit mask : 00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11|00|11
	// And op     : 00|11|00|01|00|10|00|01|00|11|00|11|00|10|00|01|00|10|00|00|00|00|00|01|00|01|00|11|00|01|00|00

	// x          : 0011000100100001001100110010000100100000000000010001001100010000
	// x << 2     : 0110001001000010011001100100001001000000000000100010011000100000
	// x | x << 2 : 0111001101100011011101110110001101100000000000110011011100110000
	// 2-bit mask : 0101010101010101010101010101010101010101010101010101010101010101
	// And op     : 0101000101000001010101010100000101000000000000010001010100010000

	// x spread   : 0101000101000001010101010100000101000000000000010001010100010000

	x := uint64(lat32)
	y := uint64(lng32)

	x = (x | (x << 16)) & 0x0000FFFF0000FFFF
	x = (x | (x << 8)) & 0x00FF00FF00FF00FF
	x = (x | (x << 4)) & 0x0F0F0F0F0F0F0F0F
	x = (x | (x << 2)) & 0x3333333333333333
	x = (x | (x << 1)) & 0x5555555555555555

	y = (y | (y << 16)) & 0x0000FFFF0000FFFF
	y = (y | (y << 8)) & 0x00FF00FF00FF00FF
	y = (y | (y << 4)) & 0x0F0F0F0F0F0F0F0F
	y = (y | (y << 2)) & 0x3333333333333333
	y = (y | (y << 1)) & 0x5555555555555555

	// The x and y 32-bit integers using the example values above are spread to the two 64-bit integers.
	// x: 0101000101000001010101010100000101000000000000010001010100010000
	// y: 0000010001010000010000010000000000000001000101010001000001010001

	// The y value (longitude) is shifted left by 1 so that longitude values are the even bit, while latitude is odd.
	// x      : 0101000101000001010101010100000101000000000000010001010100010000
	// y << 1 : 0000100010100000100000100000000000000010001010100010000010100010

	// A bitwise or operation merges the values, completing the interleave operation.
	// Or op  : 0101100111100001110101110100000101000010001010110011010110110010

	return x | (y << 1)
}

// deinterleave returns two uint32 values from a uint64 by moving every other bit to their respective values.
// To deinterleave, the spread operation that was done on the 32-bit integer during interleave is reversed.
// The spread operation moved the bits from the 32-bit integer to every other bit in a 64-bit integer.
// To retrieve every other bit, a bitwise and operation using a bitmask that alternates 0 and 1 can be used.
// This targets the even bits of the 64-bit integer.
// To retrieve odd bits, the evaluated 64-bit integer should be right shifted one before deinterleave is called.
// i.e., latitude: x, longitude: x >> 1
// The bit shift steps that were done during interleave are reversed using masks that facilitate the reverse shifts.
// While the masks are used in different order and the final mask is unique to the deinterleave process, the logic is the same.
func deinterleave(hashInt uint64) (uint32, uint32) {
	x := hashInt
	x &= 0x5555555555555555
	x = (x | (x >> 1)) & 0x3333333333333333
	x = (x | (x >> 2)) & 0x0f0f0f0f0f0f0f0f
	x = (x | (x >> 4)) & 0x00ff00ff00ff00ff
	x = (x | (x >> 8)) & 0x0000ffff0000ffff
	x = (x | (x >> 16)) & 0x00000000ffffffff

	y := hashInt >> 1
	y &= 0x5555555555555555
	y = (y | (y >> 1)) & 0x3333333333333333
	y = (y | (y >> 2)) & 0x0f0f0f0f0f0f0f0f
	y = (y | (y >> 4)) & 0x00ff00ff00ff00ff
	y = (y | (y >> 8)) & 0x0000ffff0000ffff
	y = (y | (y >> 16)) & 0x00000000ffffffff

	return uint32(x), uint32(y)
}

// encodeBitwiseOr returns a geohash string of the lat, lng coordinates based on the provided character precision.
// The decimal index for each 5-bit base32 character is generated using a bitwise or operation on the value of each bit position of a 5-bit geohash character.
func encodeBitwiseOr(lat, lng float64, precision int) string {
	latmin, latmax := -latMax, latMax
	lngmin, lngmax := -lngMax, lngMax

	idx, bit := 0, 0
	even := true

	hash := strings.Builder{}

	for hash.Len() < precision {
		if even {
			mid := (lngmin + lngmax) / 2
			if lng >= mid {
				idx |= bitPositions[bit]
				lngmin = mid
			} else {
				lngmax = mid
			}
		} else {
			mid := (latmin + latmax) / 2
			if lat >= mid {
				idx |= bitPositions[bit]
				latmin = mid
			} else {
				latmax = mid
			}
		}

		even = !even
		bit++

		if bit > 4 {
			hash.WriteByte(base32[idx])
			idx = 0
			bit = 0
		}
	}

	return hash.String()
}

// encodeDoubling returns a geohash string of the lat, lng coordinates based on the provided character precision.
// The decimal index for each 5-bit base32 character is generated using the doubling method for binary to decimal calculation.
// Reference: https://en.wikipedia.org/wiki/Double_dabble#Historical
// Reference: https://ideasawakened.com/post/double-dabble-and-conversions-from-base-10-and-base-2-number-systems
// Reference: https://www.cuemath.com/numbers/binary-to-decimal/
func encodeDoubling(lat, lng float64, precision int) string {
	latmin, latmax := -latMax, latMax
	lngmin, lngmax := -lngMax, lngMax

	idx, bit := 0, 0
	even := true

	hash := strings.Builder{}

	for hash.Len() < precision {
		if even {
			mid := (lngmin + lngmax) / 2
			if lng >= mid {
				idx = idx*2 + 1
				lngmin = mid
			} else {
				idx = idx * 2
				lngmax = mid
			}
		} else {
			mid := (latmin + latmax) / 2
			if lat >= mid {
				idx = idx*2 + 1
				latmin = mid
			} else {
				idx = idx * 2
				latmax = mid
			}
		}

		even = !even
		bit++

		if bit > 4 {
			hash.WriteByte(base32[idx])
			idx = 0
			bit = 0
		}
	}

	return hash.String()
}

// encodeStrConcat returns a geohash string of the lat, lng coordinates based on the provided character precision.
// This is a naive implemenation using string concatenation to build 5-bit binary strings.
// The 5-bit binary strings are parsed as integers that serve as the index for the base32 character.
// This function is for illustrative purposes, does not perform well, and should not be used.
func encodeStrConcat(lat, lng float64, precision int) (string, error) {
	latmin, latmax := -latMax, latMax
	lngmin, lngmax := -lngMax, lngMax

	bit := 0
	even := true

	bStr := strings.Builder{}
	hash := strings.Builder{}

	for hash.Len() < precision {
		if even {
			mid := (lngmin + lngmax) / 2
			if lng >= mid {
				bStr.WriteString("1")
				lngmin = mid
			} else {
				bStr.WriteString("0")
				lngmax = mid
			}
		} else {
			mid := (latmin + latmax) / 2
			if lat >= mid {
				bStr.WriteString("1")
				latmin = mid
			} else {
				bStr.WriteString("0")
				latmax = mid
			}
		}

		even = !even
		bit++

		if bit > 4 {
			b32idx, err := strconv.ParseInt(bStr.String(), 2, 64)
			if err != nil {
				return "", err
			}
			hash.WriteString(string(base32[b32idx]))
			bStr.Reset()
			bit = 0
		}
	}

	return hash.String(), nil
}

// validate is a helper function that checks whether the provided integer is within the specified range, returning min or max values on exception.
func validate(min, max, v int) int {
	switch {
	case v > max:
		return max
	case v < min:
		return min
	default:
		return v
	}
}
