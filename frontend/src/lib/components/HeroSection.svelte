<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";

	interface Props {
		searchValue?: string;
		hasActiveSearch?: boolean;
		resultCount?: number;
		onSearchInput?: (value: string) => void;
		onFind?: () => void;
		onOpenScanner?: () => void;
	}

	let {
		searchValue = "",
		hasActiveSearch = false,
		resultCount = 0,
		onSearchInput = () => {},
		onFind = () => {},
		onOpenScanner = () => {},
	}: Props = $props();

	function handleSubmit(event: SubmitEvent): void {
		event.preventDefault();
		onFind();
	}

</script>

<section class="mt-[60px]">
	<div class="flex w-full max-w-[816px] flex-col">
		<h1
			class="font-serif text-[42px] leading-[0.96] font-normal tracking-[-0.02em] text-[#1A1A1A] sm:text-[56px] lg:text-[70px]"
		>
			Find
			<span class="text-[#8B8B8B]">Whats Happening</span>
		</h1>

		<div class="mt-2 flex items-center gap-2 text-[21px] text-[#1A1A1A]">
			<span class="text-[#3F3F3F]">Near You, Karukutty, IN</span>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="16"
				height="16"
				viewBox="0 0 24 24"
				fill="none"
				stroke="#2B58FF"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="m18 2 4 4" />
				<path d="M2 22h4l14-14-4-4L2 18v4z" />
			</svg>
		</div>

		<form
			class="mt-7 w-full border border-[#D9D9D5] bg-white p-1"
			onsubmit={handleSubmit}
		>
			<div class="flex flex-col gap-2 sm:flex-row sm:items-center">
				<div class="flex min-w-0 flex-1 items-center gap-3 px-3">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="20"
						height="20"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="shrink-0 text-[#90908D]"
					>
						<circle cx="11" cy="11" r="8" />
						<path d="m21 21-4.3-4.3" />
					</svg>

					<Input
						type="text"
						value={searchValue}
						placeholder="Search Events, Organizers, or Topics..."
						aria-label="Search events, organizers, or topics"
						class="h-11 flex-1 !border-0 !bg-transparent !shadow-none !outline-none !ring-0 !text-base text-[#1A1A1A] placeholder:text-[#9E9E9C] focus-visible:!ring-0"
						oninput={(event) =>
							onSearchInput(
								(
									event.currentTarget as HTMLInputElement
								).value
							)}
					/>
				</div>

				<div class="flex gap-1.5 px-1 pb-1 sm:pb-0 sm:pl-0">
					<Button
						type="button"
						variant="outline"
						size="sm"
						class="!h-auto flex-1 gap-2 !rounded-[2px] border-[#D9D9D5] bg-white px-[10px] py-[3px] text-[18px] leading-[1.2] font-medium text-[#222222] sm:w-auto sm:flex-none"
						onclick={onOpenScanner}
					>
						<img
							src="/scan-qr-code.svg"
							alt=""
							class="h-4 w-4"
						/>
						Scan QR
					</Button>

					<Button
						type="submit"
						size="sm"
						class="!h-auto flex-1 !rounded-[2px] bg-[#171717] px-[10px] py-[3px] text-[18px] leading-[1.2] font-medium text-white hover:bg-[#111111] sm:w-auto sm:flex-none"
					>
						Find
					</Button>
				</div>
			</div>
		</form>
	</div>
</section>
