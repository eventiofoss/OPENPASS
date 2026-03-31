<script lang="ts">
	interface Props {
		title: string;
		viewAllHref?: string;
	}

	let { title, viewAllHref = "#" }: Props = $props();
</script>

<div class="flex items-end justify-between">
	<h2
		class="font-serif text-[44px] leading-[0.92] font-normal tracking-[-0.01em] text-[#1E1E1E] sm:text-[48px]"
	>
		{title}
	</h2>
	<a
		href={viewAllHref}
		class="text-lg font-medium text-[#1E1E1E] underline underline-offset-4 transition-colors hover:text-[#3D6B8C]"
	>
		View All
	</a>
</div>
