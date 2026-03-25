<script lang="ts">
	interface Props {
		title: string;
		viewAllHref?: string;
	}

	let { title, viewAllHref = "#" }: Props = $props();
</script>

<div class="flex items-center justify-between">
	<h2 class="font-serif text-xl sm:text-2xl">{title}</h2>
	<a
		href={viewAllHref}
		class="text-xs font-medium text-muted-foreground
			transition-colors hover:text-foreground"
	>
		View All
	</a>
</div>
