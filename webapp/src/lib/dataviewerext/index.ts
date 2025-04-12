export const SEEK_MODE_START = "START" as const;
export const SEEK_MODE_CURRENT = "CURRENT" as const;
export const SEEK_MODE_END = "END" as const;

export type DataViewExtSeekMode = "START" | "CURRENT" | "END";

export class DataViewExt {
	public dataview: DataView;
	public length: number;
	public offset: number;

	constructor(buffer?: ArrayBuffer) {
		const b = buffer ?? new ArrayBuffer(0);
		this.dataview = new DataView(b);
		this.length = b.byteLength;
		this.offset = 0;
	}

	public seek(n: number, mode: DataViewExtSeekMode) {
		const length = this.length - 1;
		const start = Math.min(length, Math.max(0, n));
		const current = Math.min(length, Math.max(0, this.offset + n));
		const end = Math.min(length, Math.max(0, length + n));
		switch (mode) {
			case "START":
				this.offset = start;
				break;
			case "CURRENT":
				this.offset = current;
				break;
			case "END":
				this.offset = end;
				break;
		}
	}

	public readBytes(n: number) {
		const result = new Uint8Array(n);
		for (let i = 0; i < n; i++) {
			result[i] = this.dataview.getUint8(this.offset + i);
		}

		this.offset += n;
		return result;
	}

	public readString(n: number) {
		let result = "";
		for (let i = this.offset; i < this.offset + n; i++) {
			result += String.fromCharCode(this.dataview.getUint8(i));
		}
		this.offset += n;
		return result;
	}

	public readInt8() {
		const result = this.dataview.getInt8(this.offset);
		this.offset += 1;
		return result;
	}

	public readUint8() {
		const result = this.dataview.getUint8(this.offset);
		this.offset += 1;
		return result;
	}

	public readInt16(le?: boolean) {
		const result = this.dataview.getInt16(this.offset, le);
		this.offset += 2;
		return result;
	}

	public readInt16BE() {
		const result = this.readInt16(false);
		return result;
	}

	public readInt16LE() {
		const result = this.readInt16(true);
		return result;
	}

	public readUint16(le?: boolean) {
		const result = this.dataview.getUint16(this.offset, le);
		this.offset += 2;
		return result;
	}

	public readUint16BE() {
		const result = this.readUint16(false);
		return result;
	}

	public readUint16LE() {
		const result = this.readUint16(true);
		return result;
	}

	public readInt32(le?: boolean) {
		const result = this.dataview.getInt32(this.offset, le);
		this.offset += 4;
		return result;
	}

	public readInt32BE() {
		const result = this.readInt32(false);
		return result;
	}

	public readInt32LE() {
		const result = this.readInt32(true);
		return result;
	}

	public readUint32(le?: boolean) {
		const result = this.dataview.getUint32(this.offset, le);
		this.offset += 4;
		return result;
	}

	public readUint32BE() {
		const result = this.readUint32(false);
		return result;
	}

	public readUint32LE() {
		const result = this.readUint32(true);
		return result;
	}

	public readUint64LE() {
		const low = this.dataview.getUint32(this.offset, true);
		const high = this.dataview.getUint32(this.offset + 4, true);
		this.offset += 8;
		return BigInt(low) + (BigInt(high) << 32n);
	}

	public readUint64BE() {
		const high = this.dataview.getUint32(this.offset, false);
		const low = this.dataview.getUint32(this.offset + 4, false);
		this.offset += 8;
		return (BigInt(high) << 32n) + BigInt(low);
	}

	// TODO: add write method
}

export const uint16ToArrayLE = (n: number) => [n & 0xff, (n >> 8) & 0xff];

export const uint16ToArrayBE = (n: number) => [(n >> 8) & 0xff, n & 0xff];

export const uint32ToArrayLE = (n: number) => [
	n & 0xff,
	(n >> 8) & 0xff,
	(n >> 16) & 0xff,
	(n >> 24) & 0xff,
];

export const uint32ToArrayBE = (n: number) => [
	(n >> 24) & 0xff,
	(n >> 16) & 0xff,
	(n >> 8) & 0xff,
	n & 0xff,
];
