<script lang="ts">
	import cardArrow from "$lib/assets/arrrow.svg";

	interface Props {
		title: string;
		imageColor?: string;
		posterUrl?: string | null;
		tags: string[];
		location?: string;
		attendees?: number;
		date?: string;
		time?: string;
		href?: string;
	}

	let {
		title,
		imageColor = "#141414",
		posterUrl,
		tags,
		location = "Venue TBA",
		attendees = 0,
		date,
		time,
		href,
	}: Props = $props();

	const DEFAULT_CARD_BACKGROUND = "#F2F2F2";
	const DEFAULT_TAG_SHADE = "#5A5A5A";
	const CARD_TAG_MAP: Record<string, string> = {
		"#CDB3FE": "#8444FD",
		"#FECEB4": "#FC8745",
		"#FEB4CA": "#FC457C",
		"#B4E1FE": "#45B5FC",
		"#FEEDB4": "#FBC404",
	};

	function normalizeHexColor(color: string): string | null {
		const raw = color.trim().replace("#", "");
		if (/^[0-9a-fA-F]{3}$/.test(raw)) {
			return `#${raw[0]}${raw[0]}${raw[1]}${raw[1]}${raw[2]}${raw[2]}`.toUpperCase();
		}
		if (/^[0-9a-fA-F]{6}$/.test(raw)) {
			return `#${raw.toUpperCase()}`;
		}
		return null;
	}

	function resolveCardBackground(color: string): string {
		return normalizeHexColor(color) ?? DEFAULT_CARD_BACKGROUND;
	}

	function resolveTagShade(cardBackground: string): string {
		return CARD_TAG_MAP[cardBackground] ?? DEFAULT_TAG_SHADE;
	}

	const cardBackground = $derived(resolveCardBackground(imageColor));
	const tagShade = $derived(resolveTagShade(cardBackground));
	const normalizedTags = $derived(
		tags
			.map((tag) => tag.replace(/^#/, "").trim())
			.filter((tag) => tag.length > 0)
	);

	function formatAttendees(value: number): string {
		if (value >= 1000000) {
			return `${(value / 1000000).toFixed(1)}m attending`;
		}
		if (value >= 1000) {
			return `${(value / 1000).toFixed(1)}k attending`;
		}
		return `${value} attending`;
	}

	const attendeeText = $derived(formatAttendees(attendees));
	const tagRow = $derived(
		normalizedTags.length > 0 ? normalizedTags : ["Tech", "Art"]
	);
</script>

<a
	href={href ?? `/events/${title.toLowerCase().replace(/\s+/g, '-').replace(/[^a-z0-9-]/g, '')}`}
	class="group block w-full max-w-[221px] rounded-[4px] border border-[#EAEAEA] p-2 pb-3"
	style="background-color: {cardBackground};"
>
	<!-- Poster -->
	<div
		class="relative h-48 w-full overflow-hidden rounded-[2px] transition-transform duration-300 group-hover:scale-[1.01]"
		style="background-color: {cardBackground};"
	>
		{#if posterUrl}
			<img
				src={posterUrl}
				alt={title}
				class="absolute inset-0 h-full w-full object-cover"
			/>
		{:else}
			<div class="absolute inset-0 bg-gradient-to-br from-[#ECECEC] via-[#DFE3EC] to-[#D9DDE6]"></div>
			<div class="absolute inset-0 flex items-center justify-center">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					width="42"
					height="42"
					viewBox="0 0 24 24"
					fill="none"
					stroke="#8A8A8A"
					stroke-width="1.6"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<rect x="3" y="3" width="18" height="18" rx="2" />
					<circle cx="9" cy="9" r="1.6" />
					<path d="m21 15-4.8-4.8L7 19.4" />
				</svg>
			</div>
		{/if}
		<div
			class="absolute inset-0 bg-gradient-to-t
				from-black/30 to-transparent"
		></div>

		{#if date}
			<div
				class="absolute bottom-3 left-3 rounded-[2px] bg-white/95 px-2.5 py-1 text-xs font-semibold text-[#202020] backdrop-blur-sm"
			>
				{date}{time ? `, ${time}` : ""}
			</div>
		{/if}
	</div>

	<!-- Metadata Row 1 -->
	<div class="mt-2 flex flex-wrap gap-1.5">
		<span class="inline-flex items-center gap-1.5 rounded-[14px] bg-[#1A1A1A] px-2 py-0.5 text-[10px] font-medium text-white">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="10"
				height="10"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<path d="M20 10c0 4.9-5.3 10-7.2 11.6a1.1 1.1 0 0 1-1.6 0C9.3 20 4 14.9 4 10a8 8 0 1 1 16 0" />
				<circle cx="12" cy="10" r="3" />
			</svg>
			{location}
		</span>
		<span class="inline-flex items-center gap-1.5 rounded-[14px] bg-[#1A1A1A] px-2 py-0.5 text-[10px] font-medium text-white">
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="10"
				height="10"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
			>
				<circle cx="9" cy="8" r="3" />
				<path d="M2 19a7 7 0 0 1 14 0" />
				<path d="M22 19a5.8 5.8 0 0 0-5-5.7" />
				<path d="M16 3.1a3 3 0 0 1 0 5.8" />
			</svg>
			{attendeeText}
		</span>
	</div>

	<!-- Metadata Row 2 -->
	<div class="mt-1.5 flex flex-wrap gap-1.5">
		{#each tagRow as tag}
			<span
				class="rounded-[2px] border px-2 py-0.5 text-[10px] font-medium text-white"
				style="background-color: {tagShade}; border-color: {tagShade};"
			>
				{tag}
			</span>
		{/each}
	</div>

	<!-- Title + Arrow -->
	<div
		class="mt-2 flex items-end justify-between gap-3"
	>
		<h3
			class="line-clamp-2 font-sans text-[16px] leading-[1.12] font-medium text-[#1E1E1E]"
		>
			{title}
		</h3>
		<img
			src={cardArrow}
			alt=""
			aria-hidden="true"
			class="h-[19px] w-[19px] shrink-0 opacity-85 transition-opacity group-hover:opacity-100"
		/>
	</div>
</a>
