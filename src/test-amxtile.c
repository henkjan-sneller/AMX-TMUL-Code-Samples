//==============================================================
// Copyright © 2022 Intel Corporation
//
// SPDX-License-Identifier: MIT
// =============================================================
#include <immintrin.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include <stdbool.h>

#define LINUX

#ifdef LINUX
#include <unistd.h> // POSIX operating system API
#include <sys/syscall.h>
#endif

#define MAX 1024
#define MAX_ROWS 16
#define MAX_COLS 64
#define STRIDE 64
#define ARCH_GET_XCOMP_PERM     0x1022
#define ARCH_REQ_XCOMP_PERM     0x1023
#define XFEATURE_XTILECFG       17
#define XFEATURE_XTILEDATA      18

//Define tile config data structure 
typedef struct __tile_config
{
    uint8_t palette_id;
    uint8_t start_row;
    uint8_t reserved_0[14];
    uint16_t colsb[8];
    uint16_t reserved_1[8];
    uint8_t rows[8];
    uint8_t reserved_2[8];
} __tilecfg;

// Initialize int8_t buffer
static void init_buffer8(int8_t* buf, int8_t value)
{
    const int rows = MAX_ROWS;
    const int colsb = MAX_COLS;

    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < colsb; j++) {
            buf[i * colsb + j] = value;
        }
    }
}

// Initialize int32_t buffer
static void init_buffer32(int32_t* buf, int32_t value)
{
    const int rows = MAX_ROWS;
    const int colsb2 = MAX_COLS / 4;

    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < colsb2; j++) {
            buf[i * colsb2 + j] = value;
        }
    }
}

// Set_tiledata_use() - Invoke syscall to set ARCH_SET_STATE_USE
static bool set_tiledata_use()
{
#ifdef LINUX
    if (syscall(SYS_arch_prctl, ARCH_REQ_XCOMP_PERM, XFEATURE_XTILEDATA)) {
        printf("\n Fail to do XFEATURE_XTILEDATA \n\n");
        return false;
    } else {
        printf("\n TILE DATA USE SET - OK \n\n");
        return true;
    }
#endif
    return true;
}

// Print int8_t buffer
static void print_buffer8(int8_t* buf, int32_t rows, int32_t colsb)
{
    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < (colsb); j++) {
            printf("%d ", buf[i * colsb + j]);
        }
        printf("\n");
    }
    printf("\n");
}

// Print int32_t buffer
static void print_buffer32(int32_t* buf, int32_t rows, int32_t colsb)
{
    for (int i = 0; i < rows; i++) {
        for (int j = 0; j < colsb; j++) {
            printf("%d ", buf[i * colsb + j]);
        }
        printf("\n");
    }
    printf("\n");
}

int main() {

    int8_t src1[MAX];
    int8_t src2[MAX];
    int32_t res[MAX / 4];
    int rows = MAX_ROWS;
    int colsb = MAX_COLS;

    // Request permission to linux kernel to run AMX 
    if (!set_tiledata_use()) {
        exit(-1);
    }

    // Create tile configuration
    __tilecfg tile_data = { 0 };
    tile_data.palette_id = 1;
    tile_data.start_row = 0;

    tile_data.rows[0] = MAX_ROWS;
    tile_data.rows[1] = MAX_ROWS;
    tile_data.rows[2] = MAX_ROWS;
    tile_data.rows[3] = MAX_ROWS;

    tile_data.colsb[0] = MAX_ROWS;
    tile_data.colsb[1] = MAX_COLS;
    tile_data.colsb[2] = MAX_COLS;
    tile_data.colsb[3] = MAX_COLS;

    // Init src matrix buffers with data
    init_buffer8(src1, 2);
    init_buffer8(src2, 2);

    print_buffer8(src1, rows, colsb);
    print_buffer8(src2, rows, colsb);

    // Init dst matrix buffers with data
    init_buffer32(res, 0);

    {  // this code is replaced by a Go function written in assembly.
        _tile_loadconfig(&tile_data); // Load tile configuration
        
        _tile_loadd(1, res, STRIDE); // Load tile rows from memory
        _tile_loadd(2, src1, STRIDE);
        _tile_loadd(3, src2, STRIDE);

        _tile_dpbssd(1, 2, 3); // Compute dot-product of bytes in tiles 
        _tile_stored(1, res, STRIDE);  // Store the tile data to memory

        // Release the tile configuration to return to the init state, 
        // which releases all storage it currently holds
        _tile_release();
    }
    printf("AMX: \n");
    print_buffer32(res, rows, colsb / 4);
}