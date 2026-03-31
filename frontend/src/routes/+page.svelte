<script lang="ts">
	import Navbar from "$lib/components/Navbar.svelte";
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

	const happeningNow = $derived(data.happeningNow || []);
	const discoverNearby = $derived(data.discoverNearby || []);

	const colors = ["#FF6B9D", "#8B5CF6", "#06B6D4", "#F59E0B", "#EC4899", "#F97316", "#22D3EE", "#3B82F6"];
	
	function getColor(index: number) {
		return colors[index % colors.length];
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
		<HeroSection />

		<!-- Happening Now -->
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
								date={event.dateLabel !== 'Date TBA' ? event.dateLabel : undefined}
								time={event.timeLabel !== 'Time TBA' ? event.timeLabel : undefined}
								href="/events/{event.slug}"
							/>
						{/each}
					</div>
				</div>
			</div>
		</section>

		<!-- Discover Nearby -->
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
								imageColor={getColor(i + happeningNow.length)}
								posterUrl={event.posterUrl}
								tags={event.tags}
								date={event.dateLabel !== 'Date TBA' ? event.dateLabel : undefined}
								time={event.timeLabel !== 'Time TBA' ? event.timeLabel : undefined}
								href="/events/{event.slug}"
							/>
						{/each}
					</div>
				</div>
			</div>
		</section>

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
</div>
