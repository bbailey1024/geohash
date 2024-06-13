## Geohash

A Go geohash library for encoding and decoding latitude and longitude coordinates to and from geohash strings or integers.

[![Go Reference](https://pkg.go.dev/badge/github.com/bbailey1024/geohash.svg)](https://pkg.go.dev/github.com/bbailey1024/geohash)
[![Go Report Card](https://goreportcard.com/badge/github.com/bbailey1024/geohash)](https://goreportcard.com/report/github.com/bbailey1024/geohash)

## About

This geohash library is largely academic and used to gain a functional understanding of the [geohash algorithm](https://en.wikipedia.org/wiki/Geohash). As such, function comments are very verbose. While this library may be suitable for import, there are other Go geohash libraries that are more mature, feature-rich, and actively maintained.

The [geohash library by mmcloughlin](https://github.com/mmcloughlin/geohash), should be considered in lieu of this library. In fact, that library was frequently referenced during development and used for validation. Ultimately, many of the functions here resemble, or are identical to, those found in mmcloughlin's library.

If interested in exploring the concepts around the geohash algorithm and its implementation, consider reviewing the Reference section below. Of note, is the [Geohash in Golang Assembly](https://mmcloughlin.com/posts/geohash-assembly) blog by Michael McLoughlin.

## Usage

### Encode

Encodes the lat, lng coordinates to a geohash string with 12 characters of precision.

    Encode(lat, lng float64) string

### EncodePrecision

Encodes the lat, lng coordinates to a geohash string of with the specified number of charater precision (1 - 12 characters). Precision that is out of bounds will be set to the min or max respectively.

    EncodePrecision(lat, lng float64, precision int) string

### EncodeHighPrecision

Encodes the lat, lng coordinates to a geohash string of with the specified number of charater precision (1 - 20 characters). Precision that is out of bounds will be set to the min or max respectively. This uses a slightly slower encoding mechanism.

    EncodeHighPrecision(lat, lng float64, precision int) string

### EncodeInt

Encodes the lat, lng coordinates to a geohash unsigned 64-bit integer with 64-bit precision.

    EncodeInt(lat, lng float64) uint64

### EncodeIntPrecision

Encodes the lat, lng coordinates to a geohash unsigned 64-bit integer with the specified bit precision (1-64 bits).

    EncodeIntPrecision(lat, lng float64) uint64

### Decoding and Conversion

The `Decode` function can be used for geohash strings with precision of 1-12 characters, while the `DecodeHighPrecision` function is required for geohash strings with precision of 13-20 characters.

When `DecodeInt` is used to decode geohash integers, a 64-bit precision is assumed. Use `DecodeIntPrecision` to specify a bit precision. Both functions will return unpredictable results if the specified precision does not match the encoding precision.

The `EncodeIntToStr` and `EncodeStrToInt` functions can convert between geohash integers and strings. The `EncodeIntToStr` function assumes the integer was generated using precision*5 bits. If this is not the case, the resulting geohash string will be malformed. A 64-bit precision integer should be right shifted 4 to generate a 60-bit precision integer to get a 12 character precision geohash string.

## References

[Wikipedia](https://en.wikipedia.org/wiki/Geohash)

[Geohash in Golang Assembly by Michael McLoughlin](https://mmcloughlin.com/posts/geohash-assembly)

[Notes on Geohashing by eugene-eeo](https://eugene-eeo.github.io/blog/geohashing.html)

[Interleave bits by Binary Magic Numbers](https://graphics.stanford.edu/~seander/bithacks.html#InterleaveBMN)

## Notable Implementations

https://github.com/mmcloughlin/geohash

https://github.com/redis/redis/blob/fc0c9c8097a5b2bc8728bec9cfee26817a702f09/src/geohash.c

https://github.com/yinqiwen/geohash-int/blob/master/geohash.c

https://github.com/chrisveness/latlon-geohash
