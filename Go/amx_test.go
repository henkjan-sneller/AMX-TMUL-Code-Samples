package amx

import (
	"fmt"
	"strings"
	"syscall"
	"testing"
)

//go:noescape
func amxTDPBSSD(dst []uint32, src1, src2 []int8, tc *TileConfig)

//go:noescape
func setMXCSR()

const MAX = 1024
const MAXROWS = 16
const MAXCOLS = 64

type TileConfig struct {
	palette   uint8 // palette selects the supported configuration of the tiles that will be used.
	startRow  uint8 // startRow is used for storing the restart values for interrupted operations.
	reserved0 [14]uint8
	colsb     [8]uint16
	reserved1 [8]uint16
	rows      [8]uint8
	reserved2 [8]uint8
}

func initBuffer8(buf []int8, value int8) {
	const rows = MAXROWS
	const colsb = MAXCOLS
	for i := 0; i < rows; i++ {
		for j := 0; j < colsb; j++ {
			buf[i*colsb+j] = int8(value)
		}
	}
}

func initBuffer32(buf []uint32, value uint32) {
	const rows = MAXROWS
	const colsb = MAXCOLS / 4
	for i := 0; i < rows; i++ {
		for j := 0; j < colsb; j++ {
			buf[i*colsb+j] = value
		}
	}
}

func amxEnableLinux() {
	const archPrctl = 0x9e          // arch_prctl - set architecture-specific thread state
	const ArchReqXcompPerm = 0x1023 // ARCH_REQ_XCOMP_PERM =  0x1023
	const xfeatureXtiledata = 0x12  // XFEATURE_XTILEDATA =  18
	r, _, err := syscall.RawSyscall6(archPrctl, ArchReqXcompPerm, xfeatureXtiledata, 0, 0, 0, 0)
	fmt.Printf("amxEnable: r=%d err=%s\n", r, err)
}

func printBuffer8(buf []int8, rows, colsb int) string {
	sb := strings.Builder{}
	sb.WriteString("\n")
	for i := 0; i < rows; i++ {
		for j := 0; j < colsb; j++ {
			sb.WriteString(fmt.Sprintf("%v, ", buf[i*colsb+j]))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func printBuffer32(buf []uint32, rows, colsb int) string {
	sb := strings.Builder{}
	sb.WriteString("\n")
	for i := 0; i < rows; i++ {
		for j := 0; j < colsb; j++ {
			sb.WriteString(fmt.Sprintf("%v, ", buf[i*colsb+j]))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func refTDPBSSD(dst []uint32, src1, src2 []int8, rows, colsb int) {
	colsb2 := rows
	for i := 0; i < rows; i++ {
		offsetC := i * colsb2
		offsetA := i * colsb
		for j := 0; j < colsb2; j++ {
			posC := offsetC + j
			value := 0
			for k := 0; k < colsb; k++ {
				posA := offsetA + k
				posB := (k * colsb2) + j
				value += int(src1[posA]) * int(src2[posB])
			}
			dst[posC] += uint32(value)
		}
	}
}

func TestTDPBSSD(t *testing.T) {
	const rows = MAXROWS
	const colsb = MAXCOLS

	src1 := make([]int8, MAX)
	src2 := make([]int8, MAX)
	dst := make([]uint32, MAX/4)

	tc := TileConfig{}
	tc.palette = 1
	tc.startRow = 0

	tc.colsb[0] = MAXROWS
	tc.colsb[1] = MAXROWS
	tc.colsb[2] = MAXROWS
	tc.colsb[3] = MAXROWS

	tc.rows[0] = MAXROWS
	tc.rows[1] = MAXCOLS
	tc.rows[2] = MAXCOLS
	tc.rows[3] = MAXCOLS

	initBuffer8(src1, 2)
	initBuffer8(src2, 2)

	t.Log(printBuffer8(src1, rows, colsb))
	t.Log(printBuffer8(src2, rows, colsb))

	setMXCSR()
	amxEnableLinux()

	initBuffer32(dst, 0)
	refTDPBSSD(dst, src1, src2, rows, colsb)
	t.Log("REF:\n" + printBuffer32(dst, rows, colsb/4))

	initBuffer32(dst, 0)
	amxTDPBSSD(dst, src1, src2, &tc)
	t.Log("AMX:\n" + printBuffer32(dst, rows, colsb/4))
}

func BenchmarkAmxTDPBSSD(b *testing.B) {

	src1 := make([]int8, MAX)
	src2 := make([]int8, MAX)
	dst := make([]uint32, MAX/4)

	tc := TileConfig{}
	tc.palette = 1
	tc.startRow = 0

	tc.colsb[0] = MAXROWS
	tc.colsb[1] = MAXROWS
	tc.colsb[2] = MAXROWS
	tc.colsb[3] = MAXROWS

	tc.rows[0] = MAXROWS
	tc.rows[1] = MAXCOLS
	tc.rows[2] = MAXCOLS
	tc.rows[3] = MAXCOLS

	initBuffer8(src1, 2)
	initBuffer8(src2, 2)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		initBuffer32(dst, 0)
		b.StartTimer()
		amxTDPBSSD(dst, src1, src2, &tc)
	}
}

func BenchmarkRefTDPBSSD(b *testing.B) {

	src1 := make([]int8, MAX)
	src2 := make([]int8, MAX)
	dst := make([]uint32, MAX/4)

	initBuffer8(src1, 2)
	initBuffer8(src2, 2)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		initBuffer32(dst, 0)
		b.StartTimer()
		refTDPBSSD(dst, src1, src2, MAXROWS, MAXCOLS)
	}
}
