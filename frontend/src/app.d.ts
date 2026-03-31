// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	interface DetectedBarcode {
		boundingBox?: DOMRectReadOnly;
		cornerPoints?: ReadonlyArray<{ x: number; y: number }>;
		format?: string;
		rawValue?: string;
	}

	interface BarcodeDetectorOptions {
		formats?: string[];
	}

	class BarcodeDetector {
		constructor(options?: BarcodeDetectorOptions);

		detect(
			image: ImageBitmapSource
		): Promise<DetectedBarcode[]>;

		static getSupportedFormats(): Promise<string[]>;
	}

	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
