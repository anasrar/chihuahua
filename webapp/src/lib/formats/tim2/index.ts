import {
	SEEK_MODE_CURRENT,
	SEEK_MODE_START,
	uint16ToArrayLE,
	uint32ToArrayLE,
	type DataViewExt,
} from "@/dataviewerext";
import type { Result } from "@/result";

export const SIGNATURE = 0x324d4954;

export type Tim2FromDataViewExtOptions = {
	offset: number;
};

export type Tim2Picture = {
	totalSize: number;
	clutSize: number;
	imageSize: number;
	headerSize: number;
	clutColors: number;
	pictureFormat: number;
	mipmapTextures: number;
	clutType: number;
	imageType: number;
	imageWidth: number;
	imageHeight: number;
	gsTex0: bigint;
	gsTex1: bigint;
	gsRegs: number;
	gsTexClut: number;
	imageData: Uint8Array<ArrayBuffer>;
	clutData: [number, number, number, number][];
};

export class Tim2 {
	public offset: number;
	public formatVersion: number;
	public formatId: number;
	public pictureTotal: number;
	public pictures: Tim2Picture[];

	constructor() {
		this.offset = 0;
		this.formatVersion = 0;
		this.formatId = 0;
		this.pictureTotal = 0;
		this.pictures = [];
	}

	private parse(dataview: DataViewExt): Result<true> {
		dataview.seek(this.offset, SEEK_MODE_START);

		const signature = dataview.readUint32LE();
		if (signature !== SIGNATURE) {
			return {
				error: new Error("TIM2 signature not match"),
			};
		}

		this.formatVersion = dataview.readUint8();
		this.formatId = dataview.readUint8();
		this.pictureTotal = dataview.readUint16LE();

		dataview.seek(8, SEEK_MODE_CURRENT);

		for (let i = 0; i < this.pictureTotal; i++) {
			const totalSize = dataview.readUint32LE();
			const clutSize = dataview.readUint32LE();
			const imageSize = dataview.readUint32LE();
			const headerSize = dataview.readUint16LE();
			const clutColors = dataview.readUint16LE();
			const pictureFormat = dataview.readUint8();
			const mipmapTextures = dataview.readUint8();
			const clutType = dataview.readUint8();
			const imageType = dataview.readUint8();
			const imageWidth = dataview.readUint16LE();
			const imageHeight = dataview.readUint16LE();
			const gsTex0 = dataview.readUint64LE();
			const gsTex1 = dataview.readUint64LE();
			const gsRegs = dataview.readUint32LE();
			const gsTexClut = dataview.readUint32LE();
			const imageData = dataview.readBytes(imageSize);
			const clutData: Tim2Picture["clutData"] = [];
			for (let i = 0; i < clutColors; i++) {
				const [r, g, b, a] = dataview.readBytes(4);
				clutData.push([r, g, b, ~~((a / 0x80) * 0xff)]);
			}

			const picture: Tim2Picture = {
				totalSize,
				clutSize,
				imageSize,
				headerSize,
				clutColors,
				pictureFormat,
				mipmapTextures,
				clutType,
				imageType,
				imageWidth,
				imageHeight,
				gsTex0,
				gsTex1,
				gsRegs,
				gsTexClut,
				imageData,
				clutData,
			};

			if (clutColors >= 32) {
				let twiddle: Tim2Picture["clutData"] = [];

				for (let i = 0; i < clutColors; i += 32) {
					twiddle = twiddle.concat(picture.clutData.slice(i, i + 8));
					twiddle = twiddle.concat(picture.clutData.slice(i + 16, i + 24));
					twiddle = twiddle.concat(picture.clutData.slice(i + 8, i + 16));
					twiddle = twiddle.concat(picture.clutData.slice(i + 24, i + 32));
				}

				picture.clutData = twiddle;
			}

			this.pictures.push(picture);
		}

		return {
			value: true,
		};
	}

	public fromDataViewExt(dataview: DataViewExt, options?: Tim2FromDataViewExtOptions) {
		this.offset = options?.offset ?? 0;
		return this.parse(dataview);
	}

	public pictureToImageData(index: number): Result<ImageData> {
		if (this.pictureTotal === 0) {
			return {
				error: new Error("Empty picture"),
			};
		}

		if (index > this.pictureTotal - 1) {
			return {
				error: new Error("Picture out of bound"),
			};
		}

		const picture = this.pictures[index];
		const data: number[] = [];
		for (const i of picture.imageData) {
			data.push(...picture.clutData[i]);
		}
		const result = new ImageData(
			new Uint8ClampedArray(data),
			picture.imageWidth,
			picture.imageHeight,
		);
		return {
			value: result,
		};
	}
}

export const timImageTypeToString = (imageType: number) => {
	switch (imageType) {
		case 0x0:
			return "None";
		case 0x01:
			return "16 Bit Color";
		case 0x02:
			return "24 Bit Color";
		case 0x03:
			return "32 Bit Color";
		case 0x04:
			return "4 Bit Texture";
		case 0x05:
			return "8 Bit Texture";
		default:
			return "Unknown";
	}
};

export const tim2FromPng = (
	bpp: string,
	width: number,
	height: number,
	indices: Uint8Array,
	palettes: number[][],
): Result<Uint8Array> => {
	const colorTotal = palettes.length;

	if (colorTotal > 256) {
		return {
			error: new Error("PNG colors exceeds the maximum allowable limit of 256"),
		};
	}

	if (bpp === "4bpp" && colorTotal > 16) {
		return {
			error: new Error("PNG colors greater than 16 can not use 4 bit perpixel"),
		};
	}

	const colors: [number, number, number, number][] = palettes.map((color) => {
		const [r, g, b, a] = color;
		const alpha = ~~(((a ?? 255) / 255) * 0x80);
		return [r, g, b, alpha];
	});

	{
		let diff = 256 - colorTotal;
		if (bpp === "4bpp") {
			diff = 16 - colorTotal;
		}
		for (let i = 0; i < diff; i++) {
			colors.push([0, 0, 0, 0]);
		}
	}

	let twiddle: Tim2Picture["clutData"] = [];
	if (bpp === "8bpp") {
		for (let i = 0; i < 256; i += 32) {
			twiddle = twiddle.concat(colors.slice(i, i + 8));
			twiddle = twiddle.concat(colors.slice(i + 16, i + 24));
			twiddle = twiddle.concat(colors.slice(i + 8, i + 16));
			twiddle = twiddle.concat(colors.slice(i + 24, i + 32));
		}
	}

	const result: number[] = [
		0x54, 0x49, 0x4d, 0x32, 0x04, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	];

	switch (bpp) {
		case "4bpp":
			result.push(...uint32ToArrayLE(64 + ~~((width * height) / 2) + 48));
			break;
		case "8bpp":
			result.push(...uint32ToArrayLE(1024 + ~~((width * height) / 2) + 48));
			break;
	}

	switch (bpp) {
		case "4bpp":
			result.push(...uint32ToArrayLE(64));
			break;
		case "8bpp":
			result.push(...uint32ToArrayLE(1024));
			break;
	}

	switch (bpp) {
		case "4bpp":
			result.push(...uint32ToArrayLE(~~((width * height) / 2)));
			break;
		case "8bpp":
			result.push(...uint32ToArrayLE(~~(width * height)));
			break;
	}

	result.push(...uint16ToArrayLE(48));

	switch (bpp) {
		case "4bpp":
			result.push(...uint16ToArrayLE(16));
			break;
		case "8bpp":
			result.push(...uint16ToArrayLE(256));
			break;
	}

	// NOTE: Picture.pict_format
	result.push(0);

	// NOTE: Picture.mipmap_textures
	result.push(1);

	// NOTE: Picture.clut_type = RGBA32|0x80
	result.push(3);

	// NOTE: Picture.image_type bpp
	switch (bpp) {
		case "4bpp":
			result.push(4);
			break;
		case "8bpp":
			result.push(5);
			break;
	}

	result.push(...uint16ToArrayLE(width));
	result.push(...uint16ToArrayLE(height));

	// NOTE: Picture.gs_tex0
	const CLD = 0;
	const CSA = 0;
	const CSM = 0;
	const CPSM = 0;
	const CBP = 0;
	const TFX = 0;
	const TCC = 0;

	const TH = Math.log2(height) | 0;
	const TW = Math.log2(width) | 0;
	let PSM = 19;
	if (bpp === "4bpp") {
		PSM = 20;
	}
	const TBW = Math.floor(width / 64);
	const TBP0 = 0;

	let gstex0 = 0n;
	gstex0 |= BigInt(CLD & 0x7) << 61n;
	gstex0 |= BigInt(CSA & 0x1f) << 56n;
	gstex0 |= BigInt(CSM & 0x1) << 55n;
	gstex0 |= BigInt(CPSM & 0xf) << 51n;
	gstex0 |= BigInt(CBP & 0x3fff) << 37n;
	gstex0 |= BigInt(TFX & 0x3) << 35n;
	gstex0 |= BigInt(TCC & 0x1) << 34n;
	gstex0 |= BigInt(TH & 0xf) << 30n;
	gstex0 |= BigInt(TW & 0xf) << 26n;
	gstex0 |= BigInt(PSM & 0x3f) << 20n;
	gstex0 |= BigInt(TBW & 0x3f) << 14n;
	gstex0 |= BigInt(TBP0 & 0x3fff);

	for (let i = 0; i < 8; i++) {
		result.push(Number((gstex0 >> BigInt(i * 8)) & 0xffn));
	}

	// NOTE: Picture.gs_tex1
	result.push(0x60, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00);

	// NOTE: Picture.gs_regs
	result.push(0x00, 0x00, 0x00, 0x00);

	// NOTE: Picture.gs_tex_clut
	result.push(0x00, 0x00, 0x00, 0x00);

	result.push(...indices);

	switch (bpp) {
		case "4bpp":
			result.push(...colors.flat());
			break;
		case "8bpp":
			result.push(...twiddle.flat());
			break;
	}

	return {
		value: new Uint8Array(result),
	};
};
