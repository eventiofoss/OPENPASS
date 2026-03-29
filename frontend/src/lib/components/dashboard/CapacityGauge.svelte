<script lang="ts">
	interface Props {
		percentage: number;
		ticketsSold: number;
		capacity: number;
	}

	let { percentage, ticketsSold, capacity }: Props = $props();

	const radius = 70;
	const stroke = 8;
	const circumference = 2 * Math.PI * radius;

	let offset = $derived(
		circumference - (percentage / 100) * circumference
	);

	let gaugeColor = $derived(
		percentage >= 90
			? "#DC2626"
			: percentage >= 70
				? "#F59E0B"
				: "#3D6B8C"
	);
</script>

<div
	class="flex flex-col items-center border border-[#141414]/10
		bg-card p-6"
>
	<p
		class="mb-5 self-start font-sans text-xs font-semibold
			uppercase tracking-[0.2em] text-[#3D6B8C]"
	>
		Capacity Utilization
	</p>

	<div class="relative">
		<svg
			width="170"
			height="170"
			viewBox="0 0 170 170"
			class="drop-shadow-sm"
		>
			<!-- Background circle -->
			<circle
				cx="85"
				cy="85"
				r={radius}
				fill="none"
				stroke="#141414"
				stroke-opacity="0.06"
				stroke-width={stroke}
			/>
			<!-- Progress arc -->
			<circle
				cx="85"
				cy="85"
				r={radius}
				fill="none"
				stroke={gaugeColor}
				stroke-width={stroke}
				stroke-linecap="square"
				stroke-dasharray={circumference}
				stroke-dashoffset={offset}
				transform="rotate(-90 85 85)"
				class="transition-all duration-700 ease-out"
			/>
		</svg>

		<div
			class="absolute inset-0 flex flex-col items-center
				justify-center"
		>
			<span
				class="font-serif text-4xl tracking-tight"
				style="color: {gaugeColor};"
			>
				{percentage.toFixed(1)}%
			</span>
		</div>
	</div>

	<div class="mt-5 flex w-full justify-between gap-4">
		<div class="text-center">
			<p
				class="font-sans text-xs font-semibold uppercase
					tracking-[0.15em] text-[#141414]/50"
			>
				Sold
			</p>
			<p class="mt-1 font-serif text-2xl text-[#141414]">
				{ticketsSold}
			</p>
		</div>

		<div
			class="w-px self-stretch bg-[#141414]/10"
		></div>

		<div class="text-center">
			<p
				class="font-sans text-xs font-semibold uppercase
					tracking-[0.15em] text-[#141414]/50"
			>
				Total
			</p>
			<p class="mt-1 font-serif text-2xl text-[#141414]">
				{capacity}
			</p>
		</div>
	</div>
</div>
