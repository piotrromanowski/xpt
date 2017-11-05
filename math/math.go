package math

import (
	"encoding/binary"
	"fmt"
	"math"
)

// rename to ibmToIEEE
func IbmToIEEE(bytes []byte) float64 {
	// B*
	//bytes := []byte{66, 42, 0, 0, 0, 0, 0, 0}
	//fmt.Println(bytes)

	bbits := binary.BigEndian.Uint64(bytes)
	fmt.Println(bbits)

	// IBM: 1-bit sign, 7-bits exponent, 56-bits mantissa

	// bitwise &
	sign := bbits & 0x8000000000000000
	exponent := (bbits & 0x7f00000000000000) >> 56
	mantissa := bbits & 0x00ffffffffffffff

	fmt.Println(sign)
	fmt.Println(exponent)
	fmt.Println(mantissa)

	if mantissa == 0 {

	}

	shift := uint64(0)
	if bbits&0x0080000000000000 > 0 {
		shift = uint64(3)
	} else if bbits&0x0040000000000000 > 0 {
		shift = uint64(2)
	} else if bbits&0x0020000000000000 > 0 {
		shift = uint64(1)
	}

	mantissa >>= shift

	// clear the 1 bit to the left of the binary point
	// this is implicit in IEEE specification
	mantissa &= 0xffefffffffffffff

	// IBM exponent is excess 64, but we subtract 65, because of the
	// implicit 1 left of the radix point for the IEEE mantissa
	exponent -= 65
	// IBM exponent is base 16, IEEE is base 2, so we multiply by 4
	exponent <<= 2
	// IEEE exponent is excess 1023, but we also increment for each
	// right-shift when aligning the mantissa's first 1-bit
	exponent = exponent + shift + 1023

	// IEEE: 1-bit sign, 11-bits exponent, 52-bits mantissa
	// We didn't shift the sign bit, so it's already in the right spot
	ieee := sign | (exponent << 52) | mantissa

	bfloat := math.Float64frombits(ieee)
	fmt.Println("float is: ")
	fmt.Println(ieee)
	fmt.Println(bfloat)
	return bfloat
}
