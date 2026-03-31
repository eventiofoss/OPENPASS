<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import SectionHeader from "./SectionHeader.svelte";

	interface Organizer {
		name: string;
		followers: string;
		events: number;
		location: string;
		avatar: string;
	}

	const organizers: Organizer[] = [
		{
			name: "Muneer S",
			followers: "10m",
			events: 256,
			location: "Ernakulam",
			avatar: "https://api.dicebear.com/9.x/personas/svg?seed=Muneer",
		},
		{
			name: "Vasudev Varma",
			followers: "2.6k",
			events: 18,
			location: "Kochi",
			avatar: "https://api.dicebear.com/9.x/personas/svg?seed=Vasudev",
		},
		{
			name: "Jasil Faras",
			followers: "2.4k",
			events: 12,
			location: "Kochi",
			avatar: "https://api.dicebear.com/9.x/personas/svg?seed=Jasil",
		},
	];
</script>

<div class="flex flex-col gap-7">
	<SectionHeader
		title="Trending Organizers"
		viewAllHref="/organizers"
	/>
	<div class="flex flex-col gap-4">
		{#each organizers as org}
			<div class="flex items-center justify-between rounded-[2px] border border-[#DDDDD8] bg-[#F8F8F6] px-3 py-2">
				<div class="flex items-center gap-3">
					<img
						src={org.avatar}
						alt={org.name}
						class="h-10 w-10 rounded-[2px] border border-[#D7D7D0] bg-white object-cover"
					/>
					<div>
						<p class="text-[24px] leading-[1.05] font-medium tracking-[-0.01em] text-[#1E1E1E]">
							{org.name}
						</p>
						<p class="mt-1 text-sm text-[#6B6B69]">
							{org.followers} followers · {org.events} events · {org.location}
						</p>
					</div>
				</div>
				<Button
					variant="default"
					size="sm"
					class="h-8 rounded-[2px] bg-[#171717] px-4 text-xs font-medium text-white hover:bg-[#0F0F0F]"
				>
					Follow
				</Button>
			</div>
		{/each}
	</div>
</div>
