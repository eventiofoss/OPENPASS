<script lang="ts">
	import type { Snippet } from "svelte";

	interface Props {
		label: string;
		value: string;
		subtitle?: string;
		icon?: Snippet;
		accentColor?: string;
	}

	let {
		label,
		value,
		subtitle,
		icon,
		accentColor = "#3D6B8C",
	}: Props = $props();
</script>

<div
	class="border border-[#141414]/10 bg-card p-6
		transition-all duration-200 hover:border-[#3D6B8C]/30
		hover:shadow-sm"
>
	<div class="flex items-start justify-between gap-3">
		<p
			class="font-sans text-xs font-semibold uppercase
				tracking-[0.2em]"
			style="color: {accentColor};"
		>
			{label}
		</p>

		{#if icon}
			<div
				class="flex h-9 w-9 shrink-0 items-center
					justify-center border border-[#141414]/8
					bg-[#141414]/3"
				style="color: {accentColor};"
			>
				{@render icon()}
			</div>
		{/if}
	</div>

	<p class="mt-3 font-serif text-4xl tracking-tight text-[#141414]">
		{value}
	</p>

	{#if subtitle}
		<p class="mt-2 font-sans text-sm text-[#141414]/60">
			{subtitle}
		</p>
	{/if}
</div>
