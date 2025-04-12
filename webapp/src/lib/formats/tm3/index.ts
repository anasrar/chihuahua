import {
	SEEK_MODE_CURRENT,
	SEEK_MODE_END,
	SEEK_MODE_START,
	type DataViewExt,
} from "@/dataviewerext";
import type { Result } from "@/result";

export const SIGNATURE = 0x00334d54;
export const ENTRY_NAME_LENGTH = 0x8;

export type Tm3Entry = {
	name: string;
	size: number;
	offset: number;
};

export type Tm3FromDataViewExtOptions = {
	offset: number;
};

export class Tm3 {
	public offset: number;
	public size: number;
	public entryTotal: number;
	public entries: Tm3Entry[];

	constructor() {
		this.offset = 0;
		this.size = 0;
		this.entryTotal = 0;
		this.entries = [];
	}

	private parse(dataview: DataViewExt): Result<true> {
		if (this.size === 0) {
			dataview.seek(0, SEEK_MODE_END);
			this.size = dataview.offset - this.offset;
		}

		dataview.seek(this.offset, SEEK_MODE_START);

		const signature = dataview.readUint32LE();
		if (signature !== SIGNATURE) {
			return {
				error: new Error("TM3 signature not match"),
			};
		}

		this.entryTotal = dataview.readUint32LE();

		dataview.seek(8, SEEK_MODE_CURRENT);

		// NOTE: get offset TIM3
		for (let i = 0; i < this.entryTotal; i++) {
			const offset = dataview.readUint32LE();
			this.entries.push({
				name: "\x00\x00\x00\x00\x00\x00\x00\x00",
				size: 0,
				offset: offset + this.offset,
			});
		}

		if ((this.entryTotal & 0x1) === 1) {
			dataview.seek(4, SEEK_MODE_CURRENT);
		}

		for (let i = 0; i < this.entryTotal; i++) {
			const name = dataview.readString(ENTRY_NAME_LENGTH);
			this.entries[i].name = name;
		}

		for (let i = 0; i < this.entryTotal; i++) {
			if (i === this.entryTotal - 1) {
				this.entries[i].size = this.offset + this.size - this.entries[i].offset;
			} else {
				this.entries[i].size = this.entries[i + 1].offset - this.entries[i].offset;
			}
		}

		return {
			value: true,
		};
	}

	public fromDataViewExt(dataview: DataViewExt, options?: Tm3FromDataViewExtOptions) {
		this.offset = options?.offset ?? 0;
		return this.parse(dataview);
	}
}
