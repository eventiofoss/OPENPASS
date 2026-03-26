<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import SectionHeader from "./SectionHeader.svelte";

	interface Organizer {
		name: string;
		followers: string;
		events: number;
		color: string;
	}

	const organizers: Organizer[] = [
		{
			name: "Muneer S",
			followers: "10m",
			events: 256,
			color: "#FF6B6B",
		},
		{
			name: "Vasudev Varma",
			followers: "24k",
			events: 18,
			color: "#4ECDC4",
		},
		{
			name: "Jasil Faras",
			followers: "14k",
			events: 12,
			color: "#FFE66D",
		},
	];
</script>

<div class="flex flex-col gap-8">
	<SectionHeader
		title="Trending Organizers"
		viewAllHref="/organizers"
	/>
	<div class="flex flex-col gap-6">
		{#each organizers as org}
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-5">
					<!-- Square avatar -->
					<div
						class="flex h-16 w-16 items-center
							justify-center text-xl
							font-bold text-white"
						style="background-color: {org.color};"
					>
						{org.name.charAt(0)}
					</div>
					<div>
						<p class="text-lg font-semibold leading-tight">
							{org.name}
						</p>
						<p class="mt-1 text-base text-muted-foreground">
							{org.followers} followers ·
							{org.events} events
						</p>
					</div>
				</div>
				<Button
					variant="outline"
					size="sm"
					class="h-10 px-6 text-sm !rounded-none"
				>
					Follow
				</Button>
			</div>
		{/each}
	</div>
</div>
