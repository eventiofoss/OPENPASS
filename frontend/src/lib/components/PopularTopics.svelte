<script lang="ts">
	import { Badge } from "$lib/components/ui/badge/index.js";
	import SectionHeader from "./SectionHeader.svelte";

	interface Topic {
		name: string;
		count: string;
	}

	const topics: Topic[] = [
		{ name: "#Tech", count: "212k" },
		{ name: "#Music", count: "40K" },
		{ name: "#IndiaFOSS", count: "50k" },
		{ name: "#ARRahman", count: "10K" },
		{ name: "#Concert", count: "10K" },
		{ name: "#federation", count: "9K" },
		{ name: "#lifo", count: "2k" },
	];
</script>

<div class="flex flex-col gap-8">
	<SectionHeader
		title="Popular Topics"
		viewAllHref="/topics"
	/>
	<div class="flex flex-wrap gap-3">
		{#each topics as topic}
			<a href="/topics/{topic.name.slice(1)}">
				<Badge
					variant="secondary"
					class="cursor-pointer border border-border
						px-6 py-3 text-lg font-medium
						transition-colors !rounded-none
						hover:bg-accent
						hover:text-accent-foreground"
				>
					{topic.name} · {topic.count}
				</Badge>
			</a>
		{/each}
	</div>
</div>
