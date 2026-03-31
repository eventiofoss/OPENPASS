<script lang="ts">
	import { page } from '$app/stores';
	import EventCard from '$lib/components/EventCard.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import Navbar from '$lib/components/Navbar.svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

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
	<Navbar user={data.user} />

	<main class="flex-1 px-6 py-8 sm:px-8 lg:px-12">
		<div class="mx-auto max-w-7xl space-y-6">
			<section class="border border-[#141414]/12 bg-card p-6 sm:p-8">
				<Badge
					variant="secondary"
					class="rounded-none border border-[#141414]/14 bg-[#141414]/3 px-3 py-1 text-[#3D6B8C]"
				>
					Event Directory
				</Badge>

				<h1 class="mt-4 font-serif text-4xl text-[#141414] sm:text-5xl">
					Discover and Join Events
				</h1>
				<p class="mt-3 max-w-3xl text-sm text-[#141414]/66 sm:text-base">
					Browse published events and open any listing to register.
				</p>
			</section>

			<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
				<a
					href={filterHref('all')}
					class="border px-4 py-3 text-center text-sm font-semibold uppercase tracking-[0.12em] transition-colors {isActive('all') ? 'border-[#3D6B8C] bg-[#3D6B8C]/8 text-[#3D6B8C]' : 'border-[#141414]/12 bg-card text-[#141414]/65 hover:border-[#141414]/24'}"
				>
					All
				</a>
				<a
					href={filterHref('happening')}
					class="border px-4 py-3 text-center text-sm font-semibold uppercase tracking-[0.12em] transition-colors {isActive('happening') ? 'border-[#3D6B8C] bg-[#3D6B8C]/8 text-[#3D6B8C]' : 'border-[#141414]/12 bg-card text-[#141414]/65 hover:border-[#141414]/24'}"
				>
					Happening
				</a>
			</div>

			{#if data.error}
				<div class="border border-[#9A3B3B]/20 bg-[#9A3B3B]/7 px-4 py-3 text-sm text-[#7A2C2C]">
					{data.error}
				</div>
			{/if}

			{#if data.events.length === 0}
				<section class="border border-dashed border-[#141414]/16 bg-card p-8 text-center">
					<p class="font-serif text-3xl text-[#141414]">No events found</p>
					<p class="mt-2 text-sm text-[#141414]/62">
						Try switching filters or check back shortly.
					</p>
				</section>
			{:else}
				<section class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
					{#each data.events as event}
						<EventCard
							title={event.title}
							imageUrl={event.posterUrl}
							imageColor="#E6EAF0"
							tags={event.tags}
							date={event.dateLabel}
							time={event.timeLabel}
							location={event.venue}
							status={event.statusLabel}
							attendeeCount={event.attendeeCount}
							href={`/events/${event.slug}`}
						/>
					{/each}
				</section>
			{/if}
		</div>
	</main>

	<Footer />
</div>
