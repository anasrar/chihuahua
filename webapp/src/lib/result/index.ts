export type Success<T> = {
	value: T;
	error?: undefined;
};

export type Err = {
	value?: undefined;
	error: Error;
};

export type Result<T> = Success<T> | Err;

export const resultTryCatch: <T>(cb: () => T) => Result<T> = (cb) => {
	try {
		const result = cb();
		return {
			value: result,
		};
	} catch (err: unknown) {
		// TODO: Parsing error
		return {
			error: err as Error,
		};
	}
};

export const resultAsyncTryCatch: <T>(cb: () => T) => Promise<Result<T>> = async (cb) => {
	try {
		const result = await cb();
		return {
			value: result,
		};
	} catch (err: unknown) {
		// TODO: Parsing error
		return {
			error: err as Error,
		};
	}
};
