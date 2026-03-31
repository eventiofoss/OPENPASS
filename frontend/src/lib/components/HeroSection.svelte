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

	function searchSummary(): string {
		if (!hasActiveSearch) {
			return "Search published events by name or scan an event QR to jump " +
				"straight to its page.";
		}

		const suffix = resultCount === 1 ? "" : "s";
		return `${resultCount} event${suffix} matching "${searchValue.trim()}".`;
	}
</script>

<section class="px-8 pb-8 pt-12 sm:px-10 lg:px-12">
	<div class="mx-auto w-full max-w-7xl">
		<h1
			class="font-serif text-6xl leading-tight
				tracking-tight sm:text-7xl lg:text-8xl"
		>
			<span class="font-normal italic">Find</span>
			<span class="font-normal text-[#8A8A8A]">
				Whats Happening
			</span>
		</h1>

		<div class="mt-2 flex items-center gap-2">
			<span class="text-lg text-muted-foreground">Near You,</span>
			<span class="text-lg font-semibold text-foreground">
				Karukutty, IN
			</span>
		</div>

		<form
			class="mt-8 border border-border bg-card p-3 shadow-sm"
			onsubmit={handleSubmit}
		>
			<div class="flex flex-col gap-3 sm:flex-row sm:items-center">
				<div class="flex min-w-0 flex-1 items-center gap-3 px-3">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="22"
						height="22"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
						class="shrink-0 text-muted-foreground"
					>
						<circle cx="11" cy="11" r="8" />
						<path d="m21 21-4.3-4.3" />
					</svg>

					<Input
						type="text"
						value={searchValue}
						placeholder="Search event names..."
						aria-label="Search event names"
						class="flex-1 !border-0 !bg-transparent
							!shadow-none !outline-none !ring-0
							!text-base focus-visible:!ring-0"
						oninput={(event) =>
							onSearchInput(
								(
									event.currentTarget as HTMLInputElement
								).value
							)}
					/>
				</div>

				<div class="flex gap-2 sm:shrink-0">
					<Button
						type="button"
						variant="outline"
						size="sm"
						class="!h-12 flex-1 gap-2 !rounded-none
							px-4 text-sm sm:w-36 sm:flex-none"
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
						class="!h-12 flex-1 !rounded-none bg-foreground
							px-4 text-sm text-background
							hover:bg-foreground/90 sm:w-36 sm:flex-none"
					>
						Find
					</Button>
				</div>
			</div>
		</form>

		<p class="mt-3 font-sans text-sm text-[#141414]/62">
			{searchSummary()}
		</p>
	</div>
</section>
