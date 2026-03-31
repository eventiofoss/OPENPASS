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

	const colors = [
		"#FF6B9D",
		"#8B5CF6",
		"#06B6D4",
		"#F59E0B",
		"#EC4899",
		"#F97316",
		"#22D3EE",
		"#3B82F6",
	];

	function getColor(index: number) {
		return colors[index % colors.length];
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

		{#if hasActiveSearch}
			<section class="px-8 py-8 sm:px-10 lg:px-12">
				<div
					bind:this={searchResultsSection}
					class="mx-auto max-w-7xl"
				>
					<div class="border border-border bg-card p-8 sm:p-12">
						<div
							class="flex flex-col gap-4 border-b border-border
								pb-6 sm:flex-row sm:items-end
								sm:justify-between"
						>
							<div>
								<p
									class="font-sans text-xs font-semibold uppercase
										tracking-[0.2em] text-[#3D6B8C]"
								>
									Search Results
								</p>
								<h2 class="mt-3 font-serif text-4xl sm:text-5xl">
									{searchResults.length}
									event{searchResults.length === 1 ? "" : "s"}
									matching "{trimmedSearch}"
								</h2>
							</div>

							<Button
								type="button"
								variant="outline"
								class="h-12 rounded-none border-[#141414]/12
									px-5 font-sans text-sm font-semibold
									uppercase tracking-[0.12em]"
								onclick={clearSearch}
							>
								Clear Search
							</Button>
						</div>

						{#if searchResults.length > 0}
							<div
								class="mt-10 grid grid-cols-2 gap-6
									sm:grid-cols-4"
							>
								{#each searchResults as event, i}
									<EventCard
										title={event.title}
										imageColor={getColor(i)}
										posterUrl={event.posterUrl}
										tags={event.tags}
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
							<div
								class="mt-10 border border-dashed
									border-[#141414]/16 bg-[#141414]/2 p-8"
							>
								<p
									class="font-serif text-3xl tracking-tight
										text-[#141414]"
								>
									No event names matched yet
								</p>
								<p
									class="mt-3 max-w-2xl font-sans text-base
										leading-7 text-[#141414]/68"
								>
									Try a shorter event name, browse all published
									events, or scan a QR code if you already have the
									event pass or invite open.
								</p>

								<div class="mt-6 flex flex-wrap gap-3">
									<Button
										type="button"
										class="h-12 rounded-none bg-[#141414]
											px-5 font-sans text-sm font-semibold
											uppercase tracking-[0.12em]
											text-white hover:bg-[#141414]/90"
										onclick={() => {
											scannerOpen = true;
										}}
									>
										Scan QR
									</Button>
									<Button
										href="/events"
										variant="outline"
										class="h-12 rounded-none border-[#141414]/12
											px-5 font-sans text-sm font-semibold
											uppercase tracking-[0.12em]"
									>
										Browse Events
									</Button>
								</div>
							</div>
						{/if}
					</div>
				</div>
			</section>
		{:else}
			<section class="px-8 py-8 sm:px-10 lg:px-12">
				<div class="mx-auto max-w-7xl">
					<div class="border border-border bg-card p-8 sm:p-12">
						<SectionHeader
							title="Happening Now"
							viewAllHref="/events?filter=happening"
						/>
						<div
							class="mt-10 grid grid-cols-2
								gap-6 sm:grid-cols-4"
						>
							{#each happeningNow as event, i}
								<EventCard
									title={event.title}
									imageColor={getColor(i)}
									posterUrl={event.posterUrl}
									tags={event.tags}
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
				</div>
			</section>

			<section class="px-8 py-14 sm:px-10 lg:px-12">
				<div class="mx-auto max-w-7xl">
					<div class="border border-border bg-card p-8 sm:p-12">
						<SectionHeader
							title="Discover Nearby"
							viewAllHref="/events?filter=nearby"
						/>
						<div
							class="mt-10 grid grid-cols-2
								gap-6 sm:grid-cols-4"
						>
							{#each discoverNearby as event, i}
								<EventCard
									title={event.title}
									imageColor={getColor(
										i + happeningNow.length
									)}
									posterUrl={event.posterUrl}
									tags={event.tags}
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
				</div>
			</section>
		{/if}

		<!-- Trending Organizers + Popular Topics -->
		<section class="px-8 py-14 sm:px-10 lg:px-12">
			<div class="mx-auto max-w-7xl">
				<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
					<div class="border border-border bg-card p-8 sm:p-12">
						<TrendingOrganizers />
					</div>
					<div class="border border-border bg-card p-8 sm:p-12">
						<PopularTopics />
					</div>
				</div>
			</div>
		</section>

		<!-- Map Section -->
		<MapSection />
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
