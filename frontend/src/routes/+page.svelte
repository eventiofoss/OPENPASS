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

	const happeningNow = [
		{
			title: "La La Land Screening Festival 2026",
			imageColor: "#FF6B9D",
			tags: ["#screening", "#LA screening"],
			date: undefined,
			time: undefined,
		},
		{
			title: "Lorem Posem Meetup",
			imageColor: "#8B5CF6",
			tags: ["#Tech", "#Registration"],
			date: undefined,
			time: undefined,
		},
		{
			title: "From Data to Destination: Wiki Workshop",
			imageColor: "#06B6D4",
			tags: ["#FreeEntry", "#LetsExplore"],
			date: undefined,
			time: undefined,
		},
		{
			title: "FOSS Meetup Kochi",
			imageColor: "#F59E0B",
			tags: ["#starting in 3hrs"],
			date: undefined,
			time: undefined,
		},
	];

	const discoverNearby = [
		{
			title: "La La Land Screening Festival 2026",
			imageColor: "#EC4899",
			tags: ["#FreeEntry", "#10 Registered"],
			date: "9 Jun",
			time: undefined,
		},
		{
			title: "Lorem Posem Meetup",
			imageColor: "#F97316",
			tags: ["#FreeEntry", "#LA happening"],
			date: "14 May, 7:00 AM",
			time: undefined,
		},
		{
			title: "From Data to Destination: Wiki Workshop",
			imageColor: "#22D3EE",
			tags: ["#FreeEntry", "#LA happening"],
			date: "14 May",
			time: undefined,
		},
		{
			title: "FOSS Meetup Kochi",
			imageColor: "#3B82F6",
			tags: ["#FreeEntry"],
			date: "12 Jul, 9:00 PM",
			time: undefined,
		},
	];
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
						{#each happeningNow as event}
							<EventCard
								title={event.title}
								imageColor={event.imageColor}
								tags={event.tags}
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
						{#each discoverNearby as event}
							<EventCard
								title={event.title}
								imageColor={event.imageColor}
								tags={event.tags}
								date={event.date}
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
