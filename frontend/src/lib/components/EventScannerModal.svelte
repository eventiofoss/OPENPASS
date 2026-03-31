<script lang="ts">
	import { goto } from "$app/navigation";
	import { onMount } from "svelte";
	import {
		Camera,
		Link2,
		RefreshCw,
		ScanLine,
		X,
	} from "@lucide/svelte";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { resolveEventPathFromQr } from "$lib/utils/event-qr";

	interface Props {
		onClose?: () => void;
	}

	let { onClose = () => {} }: Props = $props();

	let cameraFeed: HTMLVideoElement | null = null;
	let previewStream: MediaStream | null = null;
	let barcodeDetector: BarcodeDetector | null = null;
	let detectionTimer: number | null = null;
	let isDetecting = false;

	let isStarting = $state(true);
	let isNavigating = $state(false);
	let scannerMessage = $state("");
	let manualValue = $state("");
	let manualError = $state("");

	function stopScanner(): void {
		if (detectionTimer !== null) {
			window.clearInterval(detectionTimer);
			detectionTimer = null;
		}

		for (const track of previewStream?.getTracks() ?? []) {
			track.stop();
		}

		previewStream = null;
		barcodeDetector = null;

		if (cameraFeed) {
			cameraFeed.srcObject = null;
		}
	}

	function closeScanner(): void {
		stopScanner();
		onClose();
	}

	function getCameraErrorMessage(error: unknown): string {
		if (error instanceof DOMException) {
			if (error.name === "NotAllowedError") {
				return "Camera access was blocked. Allow camera permissions and " +
					"try again.";
			}

			if (error.name === "NotFoundError") {
				return "No camera was found on this device.";
			}
		}

		if (error instanceof Error && error.message.trim() !== "") {
			return error.message;
		}

		return "We could not start the camera right now.";
	}

	async function openEvent(value: string): Promise<void> {
		const path = resolveEventPathFromQr(
			value,
			window.location.origin
		);
		if (!path) {
			manualError =
				"That QR code does not point to an Eventio event yet.";
			return;
		}

		isNavigating = true;
		stopScanner();
		onClose();
		await goto(path);
	}

	async function detectFrame(): Promise<void> {
		if (
			isDetecting ||
			!barcodeDetector ||
			!cameraFeed ||
			cameraFeed.readyState < HTMLMediaElement.HAVE_CURRENT_DATA
		) {
			return;
		}

		isDetecting = true;

		try {
			const barcodes = await barcodeDetector.detect(
				cameraFeed as unknown as ImageBitmapSource
			);
			const rawValue = barcodes[0]?.rawValue?.trim();

			if (rawValue) {
				await openEvent(rawValue);
			}
		} catch {
			scannerMessage =
				"Camera is ready, but QR detection needs a clearer view.";
		} finally {
			isDetecting = false;
		}
	}

	async function startScanner(): Promise<void> {
		stopScanner();
		isStarting = true;
		scannerMessage = "";
		manualError = "";

		try {
			if (
				!navigator.mediaDevices ||
				typeof navigator.mediaDevices.getUserMedia !== "function"
			) {
				throw new Error(
					"Camera access is not available in this browser."
				);
			}

			if (typeof BarcodeDetector === "undefined") {
				scannerMessage =
					"Live QR scanning works best in Chrome or Edge right now. " +
					"Paste the event link below if needed.";
				return;
			}

			const supportedFormats: string[] =
				await BarcodeDetector.getSupportedFormats().catch(
					() => [] as string[]
				);
			if (
				supportedFormats.length > 0 &&
				!supportedFormats.includes("qr_code")
			) {
				scannerMessage =
					"This browser can open the camera, but QR decoding is not " +
					"available yet.";
				return;
			}

			barcodeDetector = new BarcodeDetector({
				formats: ["qr_code"],
			});
			previewStream = await navigator.mediaDevices.getUserMedia({
				audio: false,
				video: {
					facingMode: {
						ideal: "environment",
					},
				},
			});

			if (!cameraFeed) {
				throw new Error("Camera preview could not be created.");
			}

			cameraFeed.srcObject = previewStream;
			await cameraFeed.play();
			scannerMessage =
				"Point your camera at an event QR code to open the event page.";
			detectionTimer = window.setInterval(() => {
				void detectFrame();
			}, 300);
		} catch (error) {
			stopScanner();
			scannerMessage = getCameraErrorMessage(error);
		} finally {
			isStarting = false;
		}
	}

	async function handleManualSubmit(event: SubmitEvent): Promise<void> {
		event.preventDefault();
		manualError = "";

		if (manualValue.trim() === "") {
			manualError = "Paste an event link or slug first.";
			return;
		}

		await openEvent(manualValue);
	}

	onMount(() => {
		void startScanner();

		const onKeydown = (event: KeyboardEvent) => {
			if (event.key === "Escape" && !isNavigating) {
				closeScanner();
			}
		};

		window.addEventListener("keydown", onKeydown);

		return () => {
			window.removeEventListener("keydown", onKeydown);
			stopScanner();
		};
	});
</script>

<div class="fixed inset-0 z-50 px-4 py-6">
	<button
		type="button"
		class="absolute inset-0 bg-[#141414]/70 backdrop-blur-sm"
		aria-label="Close scanner"
		onclick={closeScanner}
	></button>

	<div class="relative mx-auto flex min-h-full max-w-6xl items-center">
		<div
			role="dialog"
			aria-modal="true"
			aria-labelledby="event-scanner-title"
			tabindex="-1"
			class="grid w-full gap-6 overflow-hidden border
				border-[#141414]/10 bg-[#F5F1E8] p-4 shadow-2xl
				lg:grid-cols-[1.15fr_0.85fr] lg:p-6"
		>
			<div class="relative overflow-hidden border border-[#141414]/12">
				<video
					bind:this={cameraFeed}
					autoplay
					muted
					playsinline
					class="aspect-[4/5] w-full bg-[#141414] object-cover"
				></video>

				<div
					class="pointer-events-none absolute inset-0 flex flex-col
						justify-between p-5"
				>
					<div class="flex items-center justify-between">
						<Badge
							variant="secondary"
							class="rounded-none border border-white/20
								bg-black/40 px-3 py-1 text-white"
						>
							<ScanLine />
							Live Scanner
						</Badge>
					</div>

					<div
						class="relative mx-auto h-[68%] w-[78%] border
							border-white/60"
					>
						<div
							class="absolute left-3 right-3 top-1/2 h-px
								animate-pulse bg-[#8FD3FF]
								shadow-[0_0_18px_rgba(143,211,255,0.9)]"
						></div>
					</div>
				</div>

				{#if isStarting}
					<div
						class="absolute inset-0 flex items-center justify-center
							bg-[#141414]/45 font-sans text-sm uppercase
							tracking-[0.2em] text-white"
					>
						Starting camera...
					</div>
				{/if}
			</div>

			<div class="flex flex-col justify-between gap-6">
				<div>
					<div class="flex items-start justify-between gap-4">
						<div>
							<p
								class="font-sans text-xs font-semibold uppercase
									tracking-[0.2em] text-[#3D6B8C]"
							>
								Event QR Scanner
							</p>
							<h2
								id="event-scanner-title"
								class="mt-3 font-serif text-4xl leading-tight
									tracking-tight text-[#141414]"
							>
								Open any event in one scan
							</h2>
						</div>

						<Button
							type="button"
							variant="outline"
							size="icon"
							class="rounded-none"
							onclick={closeScanner}
						>
							<X class="size-4" />
						</Button>
					</div>

					<p class="mt-4 font-sans text-base leading-7 text-[#141414]/70">
						Use the event QR shared on the public event page. As soon as
						we recognize it, we will take you straight there.
					</p>

					<div
						class="mt-5 border border-[#141414]/10 bg-white/70 p-4
							font-sans text-sm leading-7 text-[#141414]/72"
					>
						<div class="flex items-center gap-2">
							<Camera class="size-4 text-[#3D6B8C]" />
							<span>{scannerMessage}</span>
						</div>
					</div>
				</div>

				<form class="space-y-4" onsubmit={handleManualSubmit}>
					<div>
						<label
							for="event-link"
							class="font-sans text-xs font-semibold uppercase
								tracking-[0.2em] text-[#141414]/50"
						>
							Manual Open
						</label>
						<Input
							id="event-link"
							type="text"
							bind:value={manualValue}
							placeholder="/events/example-slug or full event URL"
							class="mt-2 h-12 rounded-none border-[#141414]/12
								bg-white px-4"
						/>
					</div>

					{#if manualError}
						<p class="font-sans text-sm text-red-600">
							{manualError}
						</p>
					{/if}

					<div class="flex flex-col gap-3 sm:flex-row">
						<Button
							type="submit"
							class="h-12 rounded-none bg-[#141414] px-5
								font-sans text-sm font-semibold uppercase
								tracking-[0.12em] text-white
								hover:bg-[#141414]/90"
							disabled={isNavigating}
						>
							<Link2 class="size-4" />
							Open Event
						</Button>

						<Button
							type="button"
							variant="outline"
							class="h-12 rounded-none border-[#141414]/12 px-5
								font-sans text-sm font-semibold uppercase
								tracking-[0.12em]"
							onclick={() => void startScanner()}
						>
							<RefreshCw class="size-4" />
							Try Camera Again
						</Button>
					</div>
				</form>
			</div>
		</div>
	</div>
</div>
