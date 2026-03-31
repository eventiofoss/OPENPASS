<script lang="ts">
	import { tick } from "svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import Navbar from "$lib/components/Navbar.svelte";
	import EventScannerModal from
		"$lib/components/EventScannerModal.svelte";
	import HeroSection from "$lib/components/HeroSection.svelte";
	import SectionHeader from "$lib/components/SectionHeader.svelte";
	import EventCard from "$lib/components/EventCard.svelte";
	import TrendingOrganizers
		from "$lib/components/TrendingOrganizers.svelte";
	import PopularTopics
		from "$lib/components/PopularTopics.svelte";
	import MapSection from "$lib/components/MapSection.svelte";
	import Footer from "$lib/components/Footer.svelte";

	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	let searchQuery = $state("");
	let scannerOpen = $state(false);
	let searchResultsSection = $state<HTMLElement | null>(null);

	const happeningNow = $derived(data.happeningNow || []);
	const discoverNearby = $derived(data.discoverNearby || []);
	const happeningNowSection = $derived(happeningNow.slice(0, 4));
	const discoverNearbySection = $derived(discoverNearby.slice(0, 4));
	const trimmedSearch = $derived(searchQuery.trim());
	const normalizedSearch = $derived(trimmedSearch.toLowerCase());
	const hasActiveSearch = $derived(trimmedSearch !== "");
	const searchResults = $derived(
		hasActiveSearch
			? discoverNearby.filter((event) =>
				event.title.toLowerCase().includes(
					normalizedSearch
				))
			: []
	);

	const searchColors = [
		"#F2F2F2",
		"#CDB3FE",
		"#FECEB4",
		"#FEB4CA",
		"#B4E1FE",
		"#FEEDB4",
		"#A882D9",
	];
	const happeningColors = [
		"#F2F2F2",
		"#FEEDB4",
		"#CDB3FE",
		"#B4E1FE",
	];
	const discoverColors = [
		"#FECEB4",
		"#FEB4CA",
		"#B4E1FE",
		"#CDB3FE",
	];

	function pickColor(palette: string[], index: number): string {
		return palette[index % palette.length];
	}

	function handleSearchInput(value: string) {
		searchQuery = value;
	}

	async function handleFind() {
		if (!hasActiveSearch) {
			return;
		}

		await tick();
		searchResultsSection?.scrollIntoView({
			behavior: "smooth",
			block: "start",
		});
	}

	function clearSearch() {
		searchQuery = "";
	}
</script>

<svelte:head>
	<title>Open Pass — Find What's Happening</title>
	<meta
		name="description"
		content="Discover events, meetups, and experiences happening near you. Open Pass is a self-hosted event registration and ticketing platform."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<main class="flex-1">
		<div class="mx-auto w-full max-w-[980px] px-5 sm:px-6">
			<!-- Hero Section -->
			<HeroSection
				searchValue={searchQuery}
				hasActiveSearch={hasActiveSearch}
				resultCount={searchResults.length}
				onSearchInput={handleSearchInput}
				onFind={handleFind}
				onOpenScanner={() => {
					scannerOpen = true;
				}}
			/>

			<div class="mt-[60px] space-y-[60px]">
				{#if hasActiveSearch}
					<section>
						<div
							bind:this={searchResultsSection}
							class="w-full border border-[#D9D9D5] bg-white p-4 sm:p-6"
						>
						<div
							class="flex flex-col gap-4 border-b border-[#D9D9D5] pb-5 sm:flex-row sm:items-end sm:justify-between"
						>
							<div>
								<p class="text-xs font-semibold uppercase tracking-[0.2em] text-[#3D6B8C]">
									Search Results
								</p>
								<h2 class="mt-2 font-serif text-[44px] leading-[0.92] font-normal tracking-[-0.01em] text-[#1E1E1E] sm:text-[48px]">
									{searchResults.length}
									event{searchResults.length === 1 ? "" : "s"}
									matching "{trimmedSearch}"
								</h2>
							</div>

							<Button
								type="button"
								variant="outline"
								class="h-10 rounded-[2px] border-[#D9D9D5] px-4 text-sm font-medium text-[#1E1E1E]"
								onclick={clearSearch}
							>
								Clear Search
							</Button>
						</div>

						{#if searchResults.length > 0}
							<div class="mt-7 grid grid-cols-1 justify-items-center gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-4 lg:justify-items-start">
								{#each searchResults as event, i}
									<EventCard
										title={event.title}
										imageColor={pickColor(searchColors, i)}
										posterUrl={event.posterUrl}
										tags={event.tags}
										location={event.venue}
										attendees={event.attendeeCount}
										date={event.dateLabel !== 'Date TBA'
											? event.dateLabel
											: undefined}
										time={event.timeLabel !== 'Time TBA'
											? event.timeLabel
											: undefined}
										href="/events/{event.slug}"
									/>
								{/each}
							</div>
						{:else}
							<div class="mt-7 border border-dashed border-[#D9D9D5] bg-[#141414]/2 p-6">
								<p class="font-serif text-4xl leading-[0.95] text-[#141414]">
									No event names matched yet
								</p>
								<p class="mt-3 max-w-2xl text-base leading-7 text-[#141414]/68">
									Try a shorter event name, browse all published events, or scan a QR code if you already have the event pass or invite open.
								</p>

								<div class="mt-6 flex flex-wrap gap-3">
									<Button
										type="button"
										class="h-10 rounded-[2px] bg-[#141414] px-4 text-sm font-medium text-white hover:bg-[#141414]/90"
										onclick={() => {
											scannerOpen = true;
										}}
									>
										Scan QR
									</Button>
									<Button
										href="/events"
										variant="outline"
										class="h-10 rounded-[2px] border-[#D9D9D5] px-4 text-sm font-medium text-[#1E1E1E]"
									>
										Browse Events
									</Button>
								</div>
							</div>
						{/if}
						</div>
					</section>
				{:else}
					<section>
						<div class="w-full border border-[#D9D9D5] bg-white p-4 sm:p-6">
						<SectionHeader
							title="Happening Now"
							viewAllHref="/events?filter=happening"
						/>
						<div class="mt-7 grid grid-cols-1 justify-items-center gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-4 lg:justify-items-start">
							{#each happeningNowSection as event, i}
								<EventCard
									title={event.title}
									imageColor={pickColor(happeningColors, i)}
									posterUrl={event.posterUrl}
									tags={event.tags}
									location={event.venue}
									attendees={event.attendeeCount}
									date={event.dateLabel !== 'Date TBA'
										? event.dateLabel
										: undefined}
									time={event.timeLabel !== 'Time TBA'
										? event.timeLabel
										: undefined}
									href="/events/{event.slug}"
								/>
							{/each}
						</div>
						</div>
					</section>

					<section>
						<div class="w-full border border-[#D9D9D5] bg-white p-4 sm:p-6">
						<SectionHeader
							title="Discover Nearby"
							viewAllHref="/events?filter=nearby"
						/>
						<div class="mt-7 grid grid-cols-1 justify-items-center gap-x-4 gap-y-8 sm:grid-cols-2 lg:grid-cols-4 lg:justify-items-start">
							{#each discoverNearbySection as event, i}
								<EventCard
									title={event.title}
									imageColor={pickColor(discoverColors, i)}
									posterUrl={event.posterUrl}
									tags={event.tags}
									location={event.venue}
									attendees={event.attendeeCount}
									date={event.dateLabel !== 'Date TBA'
										? event.dateLabel
										: undefined}
									time={event.timeLabel !== 'Time TBA'
										? event.timeLabel
										: undefined}
									href="/events/{event.slug}"
								/>
							{/each}
						</div>
						</div>
					</section>
				{/if}

				<!-- Trending Organizers + Popular Topics -->
				<section>
					<div class="w-full grid grid-cols-1 gap-6 lg:grid-cols-2">
						<div class="border border-[#D9D9D5] bg-white p-4 sm:p-6">
							<TrendingOrganizers />
						</div>
						<div class="border border-[#D9D9D5] bg-white p-4 sm:p-6">
							<PopularTopics />
						</div>
					</div>
				</section>

				<!-- Map Section -->
				<MapSection />
			</div>
		</div>
	</main>

	<Footer />

	{#if scannerOpen}
		<EventScannerModal
			onClose={() => {
				scannerOpen = false;
			}}
		/>
	{/if}
</div>
