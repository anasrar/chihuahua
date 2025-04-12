const mem = new Uint8Array(1024 * 1024 * 4);

// prettier-ignore
const block32 = [
  0, 1, 4, 5, 16, 17, 20, 21, 2, 3, 6,
  7, 18, 19, 22, 23, 8, 9, 12, 13, 24, 25,
  28, 29, 10, 11, 14, 15, 26, 27, 30, 31,
];

const columnWord32 = [0, 1, 4, 5, 8, 9, 12, 13, 2, 3, 6, 7, 10, 11, 14, 15];

const writeTexPSMCT32 = (
	dbp: number,
	dbw: number,
	dsax: number,
	dsay: number,
	rrw: number,
	rrh: number,
	data: number[],
) => {
	let src = 0;
	const startBlockPos = dbp * 64;

	for (let y = dsay; y < dsay + rrh; y++) {
		for (let x = dsax; x < dsax + rrw; x++) {
			const pageX = ~~(x / 64);
			const pageY = ~~(y / 32);
			const page = pageX + pageY * dbw;

			const px = x - pageX * 64;
			const py = y - pageY * 32;

			const blockX = ~~(px / 8);
			const blockY = ~~(py / 8);
			const block = block32[blockX + blockY * 8];

			const bx = px - blockX * 8;
			const by = py - blockY * 8;

			const column = ~~(by / 2);

			const cx = bx;
			const cy = by - column * 2;
			const cw = columnWord32[cx + cy * 8];

			const pos = (startBlockPos + page * 2048 + block * 64 + column * 16 + cw) * 4;

			mem[pos] = data[src];
			mem[pos + 1] = data[src + 1];
			mem[pos + 2] = data[src + 2];
			mem[pos + 3] = data[src + 3];
			src += 4;
		}
	}
};

const readTexPSMCT32 = (
	dbp: number,
	dbw: number,
	dsax: number,
	dsay: number,
	rrw: number,
	rrh: number,
	data: number[],
) => {
	let src = 0;
	const startBlockPos = dbp * 64;

	for (let y = dsay; y < dsay + rrh; y++) {
		for (let x = dsax; x < dsax + rrw; x++) {
			const pageX = ~~(x / 64);
			const pageY = ~~(y / 32);
			const page = pageX + pageY * dbw;

			const px = x - pageX * 64;
			const py = y - pageY * 32;

			const blockX = ~~(px / 8);
			const blockY = ~~(py / 8);
			const block = block32[blockX + blockY * 8];

			const bx = px - blockX * 8;
			const by = py - blockY * 8;

			const column = ~~(by / 2);

			const cx = bx;
			const cy = by - column * 2;
			const cw = columnWord32[cx + cy * 8];

			const pos = (startBlockPos + page * 2048 + block * 64 + column * 16 + cw) * 4;

			data[src] = mem[pos];
			data[src + 1] = mem[pos + 1];
			data[src + 2] = mem[pos + 2];
			data[src + 3] = mem[pos + 3];
			src += 4;
		}
	}
};

// prettier-ignore
const block8 = [
  0, 1, 4, 5, 16, 17, 20, 21, 2, 3, 6, 7, 18, 19, 22, 23,
  8, 9, 12, 13, 24, 25, 28, 29, 10, 11, 14, 15, 26, 27, 30, 31,
];

// prettier-ignore
const columnWord8: number[][] = [
  [
    0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
    2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
    8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
    10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
  ],
  [
    8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
    10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
    0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
    2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
  ],
];

// prettier-ignore
const columnByte8 = [
  0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2,
  1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3,
];

const writeTexPSMT8 = (
	dbp: number,
	dbw: number,
	dsax: number,
	dsay: number,
	rrw: number,
	rrh: number,
	data: number[],
) => {
	dbw >>= 1;
	let src = 0;
	const startBlockPos = dbp * 64;

	for (let y = dsay; y < dsay + rrh; y++) {
		for (let x = dsax; x < dsax + rrw; x++) {
			const pageX = ~~(x / 128);
			const pageY = ~~(y / 64);
			const page = pageX + pageY * dbw;

			const px = x - pageX * 128;
			const py = y - pageY * 64;

			const blockX = ~~(px / 16);
			const blockY = ~~(py / 16);
			const block = block8[blockX + blockY * 8];

			const bx = px - blockX * 16;
			const by = py - blockY * 16;

			const column = ~~(by / 4);

			const cx = bx;
			const cy = by - column * 4;
			const cw = columnWord8[column & 1][cx + cy * 16];
			const cb = columnByte8[cx + cy * 16];

			const pos = startBlockPos + page * 2048 + block * 64 + column * 16 + cw;

			mem[pos * 4 + cb] = data[src];

			src++;
		}
	}
};

const readTexPSMT8 = (
	dbp: number,
	dbw: number,
	dsax: number,
	dsay: number,
	rrw: number,
	rrh: number,
	data: number[],
) => {
	dbw >>= 1;
	let src = 0;
	const startBlockPos = dbp * 64;

	for (let y = dsay; y < dsay + rrh; y++) {
		for (let x = dsax; x < dsax + rrw; x++) {
			const pageX = ~~(x / 128);
			const pageY = ~~(y / 64);
			const page = pageX + pageY * dbw;

			const px = x - pageX * 128;
			const py = y - pageY * 64;

			const blockX = ~~(px / 16);
			const blockY = ~~(py / 16);
			const block = block8[blockX + blockY * 8];

			const bx = px - blockX * 16;
			const by = py - blockY * 16;

			const column = ~~(by / 4);

			const cx = bx;
			const cy = by - column * 4;
			const cw = columnWord8[column & 1][cx + cy * 16];
			const cb = columnByte8[cx + cy * 16];

			const pos = startBlockPos + page * 2048 + block * 64 + column * 16 + cw;

			data[src] = mem[pos * 4 + cb];

			src++;
		}
	}
};

// prettier-ignore
const block4 = [
  0, 2, 8, 10, 1, 3, 9, 11, 4, 6, 12,
  14, 5, 7, 13, 15, 16, 18, 24, 26, 17, 19,
  25, 27, 20, 22, 28, 30, 21, 23, 29, 31,
];

// prettier-ignore
const columnWord4: number[][] = [
  [
    0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
    0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
    2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
    2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
    8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
    8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
    10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
    10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
  ],
  [
    8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
    8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5,
    10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
    10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7,
    0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
    0, 1, 4, 5, 8, 9, 12, 13, 0, 1, 4, 5, 8, 9, 12, 13,
    2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
    2, 3, 6, 7, 10, 11, 14, 15, 2, 3, 6, 7, 10, 11, 14, 15,
  ],
];

// prettier-ignore
const columnByte4 = [
  0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2, 2, 2, 2, 2, 4, 4, 4, 4, 4, 4,
  4, 4, 6, 6, 6, 6, 6, 6, 6, 6, 0, 0, 0, 0, 0, 0, 0, 0, 2, 2, 2, 2,
  2, 2, 2, 2, 4, 4, 4, 4, 4, 4, 4, 4, 6, 6, 6, 6, 6, 6, 6, 6,
  1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3, 3, 3, 3, 3, 5, 5, 5, 5, 5, 5,
  5, 5, 7, 7, 7, 7, 7, 7, 7, 7, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 3,
  3, 3, 3, 3, 5, 5, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 7, 7, 7,
];

const writeTexPSMT4 = (
	dbp: number,
	dbw: number,
	dsax: number,
	dsay: number,
	rrw: number,
	rrh: number,
	data: number[],
) => {
	dbw >>= 1;
	let src = 0;
	const startBlockPos = dbp * 64;

	let odd = false;

	for (let y = dsay; y < dsay + rrh; y++) {
		for (let x = dsax; x < dsax + rrw; x++) {
			const pageX = ~~(x / 128);
			const pageY = ~~(y / 128);
			const page = pageX + pageY * dbw;

			const px = x - pageX * 128;
			const py = y - pageY * 128;

			const blockX = ~~(px / 32);
			const blockY = ~~(py / 16);
			const block = block4[blockX + blockY * 4];

			const bx = px - blockX * 32;
			const by = py - blockY * 16;

			const column = ~~(by / 4);

			const cx = bx;
			const cy = by - column * 4;
			const cw = columnWord4[column & 1][cx + cy * 32];
			const cb = columnByte4[cx + cy * 32];

			const pos = (startBlockPos + page * 2048 + block * 64 + column * 16 + cw) * 4 + (cb >> 1);

			if ((cb & 1) != 0) {
				if (odd) {
					mem[pos] = (mem[pos] & 0x0f) | (data[src] & 0xf0);
				} else {
					mem[pos] = (mem[pos] & 0x0f) | ((data[src] << 4) & 0xf0);
				}
			} else {
				if (odd) {
					mem[pos] = (mem[pos] & 0xf0) | ((data[src] >> 4) & 0x0f);
				} else {
					mem[pos] = (mem[pos] & 0xf0) | (data[src] & 0x0f);
				}
			}

			if (odd) {
				src++;
			}

			odd = !odd;
		}
	}
};

const readTexPSMT4 = (
	dbp: number,
	dbw: number,
	dsax: number,
	dsay: number,
	rrw: number,
	rrh: number,
	data: number[],
) => {
	dbw >>= 1;
	let src = 0;
	const startBlockPos = dbp * 64;

	let odd = false;

	for (let y = dsay; y < dsay + rrh; y++) {
		for (let x = dsax; x < dsax + rrw; x++) {
			const pageX = ~~(x / 128);
			const pageY = ~~(y / 128);
			const page = pageX + pageY * dbw;

			const px = x - pageX * 128;
			const py = y - pageY * 128;

			const blockX = ~~(px / 32);
			const blockY = ~~(py / 16);
			const block = block4[blockX + blockY * 4];

			const bx = px - blockX * 32;
			const by = py - blockY * 16;

			const column = ~~(by / 4);

			const cx = bx;
			const cy = by - column * 4;
			const cw = columnWord4[column & 1][cx + cy * 32];
			const cb = columnByte4[cx + cy * 32];

			const pos = startBlockPos + page * 2048 + block * 64 + column * 16 + cw;

			const pix = mem[pos * 4 + (cb >> 1)];

			if ((cb & 1) != 0) {
				if (odd) {
					data[src] = (data[src] & 0x0f) | (pix & 0xf0);
				} else {
					data[src] = (data[src] & 0xf0) | ((pix >> 4) & 0x0f);
				}
			} else {
				if (odd) {
					data[src] = (data[src] & 0x0f) | ((pix << 4) & 0xf0);
				} else {
					data[src] = (data[src] & 0xf0) | (pix & 0x0f);
				}
			}

			if (odd) {
				src++;
			}

			odd = !odd;
		}
	}
};

export const unswizzle4 = (data: number[], width: number, height: number) => {
	const result: number[] = new Array(data.length);
	const rrw = ~~(width / 2);
	const rrh = ~~(height / 4);
	writeTexPSMCT32(0, ~~(rrw / 64), 0, 0, rrw, rrh, data);
	readTexPSMT4(0, ~~(width / 64), 0, 0, width, height, result);
	return result;
};

export const swizzle4 = (data: number[], width: number, height: number) => {
	const result: number[] = new Array(data.length);
	const rrw = ~~(width / 2);
	const rrh = ~~(height / 4);
	writeTexPSMT4(0, ~~(width / 64), 0, 0, width, height, data);
	readTexPSMCT32(0, ~~(rrw / 64), 0, 0, rrw, rrh, result);
	return result;
};

export const unswizzle8 = (data: number[], width: number, height: number) => {
	const result: number[] = new Array(data.length);
	const rrw = ~~(width / 2);
	const rrh = ~~(height / 2);
	writeTexPSMCT32(0, ~~(rrw / 64), 0, 0, rrw, rrh, data);
	readTexPSMT8(0, ~~(width / 64), 0, 0, width, height, result);
	return result;
};

export const swizzle8 = (data: number[], width: number, height: number) => {
	const result: number[] = new Array(data.length);
	const rrw = ~~(width / 2);
	const rrh = ~~(height / 2);
	writeTexPSMT8(0, ~~(width / 64), 0, 0, width, height, data);
	readTexPSMCT32(0, ~~(rrw / 64), 0, 0, rrw, rrh, result);
	return result;
};
