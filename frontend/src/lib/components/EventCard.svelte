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
	class="group block min-w-[200px] max-w-[220px]
		shrink-0 sm:min-w-[220px]"
>
	<!-- Image Placeholder -->
	<div
		class="relative aspect-[4/3] w-full overflow-hidden
			rounded-lg transition-transform duration-300
			group-hover:scale-[1.02]"
		style="background-color: {imageColor};"
	>
		<!-- Gradient overlay for visual depth -->
		<div
			class="absolute inset-0 bg-gradient-to-t
				from-black/20 to-transparent"
		></div>

		<!-- Date/time badge -->
		{#if date}
			<div
				class="absolute bottom-2 left-2 rounded
					bg-white/90 px-1.5 py-0.5 text-[10px]
					font-medium text-foreground backdrop-blur-sm"
			>
				{date}{time ? `, ${time}` : ""}
			</div>
		{/if}
	</div>

	<!-- Tags -->
	<div class="mt-2 flex flex-wrap gap-1">
		{#each tags as tag}
			<Badge
				variant="secondary"
				class="rounded-full px-2 py-0 text-[10px]
					font-medium"
			>
				{tag}
			</Badge>
		{/each}
	</div>

	<!-- Title + Bookmark -->
	<div class="mt-1.5 flex items-start justify-between gap-2">
		<h3
			class="line-clamp-2 text-sm font-medium
				leading-tight"
		>
			{title}
		</h3>
		<button
			class="shrink-0 pt-0.5 text-muted-foreground
				transition-colors hover:text-foreground"
			aria-label="Bookmark event"
			onclick={(e) => e.preventDefault()}
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="14"
				height="14"
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
