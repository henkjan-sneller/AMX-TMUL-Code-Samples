#include "textflag.h"
#include "funcdata.h"
#include "go_asm.h"

TEXT ·setMXCSR(SB),NOSPLIT,$8
  VSTMXCSR 0(SP)
  ORL   $0x8040, 0(SP)
  VLDMXCSR 0(SP)
RET

// func amxTDPBSSD(dst []uint32, src1, src2 []int8, tc *TileConfig)
TEXT ·amxTDPBSSD(SB), NOSPLIT|NOFRAME, $0

    MOVL $64, CX // set STRIDE to 64

    MOVQ tc+72(FP),DX
	LONG $0x4978e2c4; BYTE $0x02                            // LDTILECFG [RDX]

    MOVQ dst_base+0(FP),DX
	LONG $0x4b7be2c4; WORD $0x0a04                          // TILELOADD TMM0, [RDX+RCX*1]

    MOVQ src1_base+24(FP),DX
	LONG $0x4b7be2c4; WORD $0x0a0c                          // TILELOADD TMM1, [RDX+RCX*1]

    MOVQ src2_base+48(FP),DX
	LONG $0x4b7be2c4; WORD $0x0a14                          // TILELOADD TMM2, [RDX+RCX*1]

    // Compute dot-product of bytes in tiles
	LONG $0x5e6be2c4; BYTE $0xc1                            // TDPBSSD TMM0, TMM1, TMM2

    // Store the tile data to memory
    MOVQ dst_base+0(FP),DX
	LONG $0x4b7ae2c4; WORD $0x0a04                          // TILESTORED [RDX+RCX*1], TMM0

	LONG $0x4978e2c4; BYTE $0xc0                            // TILERELEASE
RET
