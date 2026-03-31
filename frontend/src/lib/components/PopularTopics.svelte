<script lang="ts">
	import SectionHeader from "./SectionHeader.svelte";

	interface Topic {
		name: string;
		count: string;
	}

	const topics: Topic[] = [
		{ name: "Tech", count: "212K" },
		{ name: "Music", count: "40K" },
		{ name: "IndiaFOSS", count: "50K" },
		{ name: "ARRahman", count: "10K" },
		{ name: "Concert", count: "10K" },
		{ name: "Federation", count: "9K" },
		{ name: "Life", count: "2K" },
	];
</script>

<div class="flex flex-col gap-8">
	<SectionHeader
		title="Popular Topics"
		viewAllHref="/topics"
	/>
	<div class="flex flex-wrap gap-3">
		{#each topics as topic}
			<a href="/topics/{topic.name.toLowerCase()}">
				<span
					class="inline-flex font-sans rounded-[2px] border border-[#8A8A8A] bg-[#F2F2F2] px-[10px] py-[7px] text-[20px] font-medium leading-none text-[#1E1E1E]"
				>
					#{topic.name} · {topic.count}
				</span>
			</a>
		{/each}
	</div>
</div>
