<script lang="ts">
	interface Props {
		title: string;
		viewAllHref?: string;
	}

	let { title, viewAllHref = "#" }: Props = $props();
</script>

<div class="flex items-end justify-between">
	<h2
		class="font-serif text-4xl sm:text-5xl"
	>
		{title}
	</h2>
	<a
		href={viewAllHref}
		class="text-lg font-medium text-foreground
			underline underline-offset-4
			transition-colors hover:text-accent"
	>
		View All
	</a>
</div>
