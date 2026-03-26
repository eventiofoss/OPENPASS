<script lang="ts">
	import { Badge } from "$lib/components/ui/badge/index.js";

	interface Props {
		title: string;
		imageColor: string;
		tags: string[];
		date?: string;
		time?: string;
	}

	let {
		title,
		imageColor,
		tags,
		date,
		time,
	}: Props = $props();
</script>

<a
	href="/events/{title.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '')}"
	class="group block w-full min-w-[240px] flex-1
		shrink-0"
>
	<!-- Image -->
	<div
		class="relative aspect-[4/3] w-full overflow-hidden
			transition-transform duration-300
			group-hover:scale-[1.01]"
		style="background-color: {imageColor};"
	>
		<div
			class="absolute inset-0 bg-gradient-to-t
				from-black/30 to-transparent"
		></div>

		{#if date}
			<div
				class="absolute bottom-4 left-4
					bg-white/90 px-3 py-1.5 text-sm
					font-semibold text-foreground
					backdrop-blur-sm"
			>
				{date}{time ? `, ${time}` : ""}
			</div>
		{/if}
	</div>

	<!-- Tags -->
	<div class="mt-4 flex flex-wrap gap-2">
		{#each tags as tag}
			<Badge
				variant="secondary"
				class="px-3 py-1 text-sm font-medium
					!rounded-none"
			>
				{tag}
			</Badge>
		{/each}
	</div>

	<!-- Title + Bookmark -->
	<div
		class="mt-3 flex items-start justify-between
			gap-3"
	>
		<h3
			class="line-clamp-2 text-lg font-semibold
				leading-snug"
		>
			{title}
		</h3>
		<button
			class="shrink-0 pt-1
				text-muted-foreground
				transition-colors hover:text-foreground"
			aria-label="Bookmark event"
			onclick={(e) => e.preventDefault()}
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="22"
				height="22"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path
					d="m19 21-7-4-7 4V5a2 2 0
						0 1 2-2h10a2 2 0 0 1 2
						2v16z"
				/>
			</svg>
		</button>
	</div>
</a>
