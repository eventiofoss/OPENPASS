<script lang="ts">
	import { Badge } from "$lib/components/ui/badge/index.js";
	import {
		LayoutDashboard,
		Plus,
		CalendarDays,
	} from "@lucide/svelte";
	import type { OrganizerEvent } from "$lib/api/analytics";

	interface Props {
		events: OrganizerEvent[];
		activeEventId?: string;
	}

	let { events, activeEventId }: Props = $props();

	function statusColor(
		status: string
	): string {
		switch (status) {
			case "active":
				return "bg-emerald-500/12 text-emerald-700";
			case "full":
				return "bg-amber-500/12 text-amber-700";
			case "cancelled":
				return "bg-red-500/12 text-red-700";
			default:
				return "bg-[#141414]/6 text-[#141414]/60";
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);

		if (Number.isNaN(date.getTime())) {
			return "";
		}

		return date.toLocaleDateString("en-IN", {
			day: "numeric",
			month: "short",
		});
	}
</script>

<aside
	class="flex h-full flex-col border border-[#141414]/10
		bg-card"
>
	<!-- Header -->
	<div class="border-b border-[#141414]/10 px-5 py-5">
		<div class="flex items-center gap-2">
			<LayoutDashboard
				class="size-4 text-[#3D6B8C]"
			/>
			<p
				class="font-sans text-xs font-semibold uppercase
					tracking-[0.25em] text-[#3D6B8C]"
			>
				Dashboard
			</p>
		</div>
	</div>

	<!-- Event list -->
	<nav
		class="flex-1 space-y-1 overflow-y-auto px-3 py-3
			no-scrollbar"
	>
		{#if events.length === 0}
			<div
				class="mx-1 border border-dashed border-[#141414]/14
					bg-[#141414]/2 p-4"
			>
				<p class="font-sans text-sm font-medium text-[#141414]">
					No events yet
				</p>
				<p class="mt-2 font-sans text-xs leading-6 text-[#141414]/60">
					Create your first event to unlock analytics and attendee tools.
				</p>
			</div>
		{:else}
			{#each events as event (event.id)}
				{@const isActive = event.id === activeEventId}
				<a
					href="/organizer/dashboard/{event.id}"
					class="group flex flex-col gap-1.5 border px-4
						py-3.5 transition-colors
						{isActive
						? 'border-[#3D6B8C]/25 bg-[#3D6B8C]/6'
						: 'border-transparent hover:border-[#141414]/8 hover:bg-[#141414]/2'}"
				>
					<p
						class="line-clamp-1 font-sans text-sm
							font-medium leading-snug
							{isActive
							? 'text-[#141414]'
							: 'text-[#141414]/78'}"
					>
						{event.title}
					</p>

					<div class="flex items-center gap-2">
						<Badge
							variant="secondary"
							class="rounded-none px-2 py-0 text-[10px]
								font-semibold uppercase
								tracking-[0.1em]
								{statusColor(event.status)}"
						>
							{event.status}
						</Badge>

						{#if event.start_date}
							<span
								class="flex items-center gap-1 text-xs
									text-[#141414]/45"
							>
								<CalendarDays class="size-3" />
								{formatDate(event.start_date)}
							</span>
						{/if}
					</div>
				</a>
			{/each}
		{/if}
	</nav>

	<!-- Create new -->
	<div class="border-t border-[#141414]/10 px-4 py-4">
		<a
			href="/organizer/dashboard/events/new"
			class="flex items-center justify-center gap-2
				border border-[#141414]/12 bg-[#141414]/3
				px-4 py-3 font-sans text-sm font-semibold
				text-[#141414]/70 transition-colors
				hover:border-[#3D6B8C] hover:text-[#3D6B8C]"
		>
			<Plus class="size-4" />
			New Event
		</a>
	</div>
</aside>
