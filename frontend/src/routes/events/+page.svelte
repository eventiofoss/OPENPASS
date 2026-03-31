<script lang="ts">
	import { page } from '$app/stores';
	import EventCard from '$lib/components/EventCard.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const eventCardColors = [
		'#F2F2F2',
		'#CDB3FE',
		'#FECEB4',
		'#FEB4CA',
		'#B4E1FE',
		'#FEEDB4',
		'#A882D9'
	];

	function pickCardColor(index: number): string {
		return eventCardColors[index % eventCardColors.length];
	}

	function filterHref(filter: 'all' | 'happening'): string {
		const params = new URLSearchParams($page.url.searchParams);
		params.set('filter', filter);
		const query = params.toString();
		return query ? `/events?${query}` : '/events';
	}

	function isActive(filter: 'all' | 'happening'): boolean {
		return data.filter === filter;
	}
</script>

<svelte:head>
	<title>Events | Open Pass</title>
	<meta
		name="description"
		content="Discover and join published events on Open Pass."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<main class="flex-1">
		<div class="mx-auto w-full max-w-[980px] px-5 pb-[60px] pt-[60px] sm:px-6">
			<section>
				<Badge
					variant="secondary"
					class="font-sans rounded-[2px] border border-[#141414]/14 bg-[#141414]/3 px-3 py-1 text-[#3D6B8C]"
				>
					Event Directory
				</Badge>

				<h1 class="mt-4 font-serif text-4xl leading-[0.94] text-[#141414] sm:text-5xl">
					Discover and Join Events
				</h1>
				<p class="mt-3 max-w-3xl font-sans text-sm text-[#141414]/66 sm:text-base">
					Browse published events and open any listing to register.
				</p>
			</section>

			<div class="mt-8 w-full rounded-none border border-[#E6E6E6] bg-white p-8 shadow-sm">
				<nav class="flex items-center gap-8">
					<a
						href={filterHref('all')}
						class="font-sans text-sm tracking-[0.08em] uppercase pb-2 transition-colors {isActive('all') ? 'text-[#141414] font-bold border-b-2 border-[#141414]' : 'text-[#8A8A8A] font-medium border-b-2 border-transparent hover:text-[#4D4D4D]'}"
					>
						All
					</a>
					<a
						href={filterHref('happening')}
						class="font-sans text-sm tracking-[0.08em] uppercase pb-2 transition-colors {isActive('happening') ? 'text-[#141414] font-bold border-b-2 border-[#141414]' : 'text-[#8A8A8A] font-medium border-b-2 border-transparent hover:text-[#4D4D4D]'}"
					>
						Happening
					</a>
				</nav>

				{#if data.error}
					<div class="mt-8 border border-[#9A3B3B]/20 bg-[#9A3B3B]/7 px-4 py-3 text-sm text-[#7A2C2C]">
						{data.error}
					</div>
				{/if}

				{#if data.events.length === 0}
					<section class="mt-8 border border-dashed border-[#141414]/16 bg-card p-8 text-center">
						<p class="font-serif text-3xl text-[#141414]">No events found</p>
						<p class="mt-2 font-sans text-sm text-[#141414]/62">
							Try switching filters or check back shortly.
						</p>
					</section>
				{:else}
					<section class="mt-8 grid grid-cols-1 gap-6 justify-items-start md:grid-cols-2 lg:grid-cols-4">
						{#each data.events as event, i}
							<EventCard
								title={event.title}
								posterUrl={event.posterUrl}
								imageColor={pickCardColor(i)}
								tags={event.tags}
								location={event.venue}
								attendees={event.attendeeCount}
								date={event.dateLabel !== 'Date TBA' ? event.dateLabel : undefined}
								time={event.timeLabel !== 'Time TBA' ? event.timeLabel : undefined}
								href={`/events/${event.slug}`}
							/>
						{/each}
					</section>
				{/if}
			</div>
		</div>
	</main>

	<Footer />
</div>
