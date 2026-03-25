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

<div class="flex flex-col gap-4">
	<SectionHeader
		title="Trending Organizers"
		viewAllHref="/organizers"
	/>
	<div class="flex flex-col gap-3">
		{#each organizers as org}
			<div
				class="flex items-center justify-between"
			>
				<div class="flex items-center gap-3">
					<!-- Avatar placeholder -->
					<div
						class="flex h-10 w-10 items-center
							justify-center rounded-full
							text-sm font-bold text-white"
						style="background-color: {org.color};"
					>
						{org.name.charAt(0)}
					</div>
					<div>
						<p
							class="text-sm font-medium
								leading-tight"
						>
							{org.name}
						</p>
						<p
							class="text-xs
								text-muted-foreground"
						>
							{org.followers} followers · {org.events}
							events
						</p>
					</div>
				</div>
				<Button
					variant="outline"
					size="sm"
					class="h-7 rounded-full px-4 text-xs"
				>
					Follow
				</Button>
			</div>
		{/each}
	</div>
</div>
