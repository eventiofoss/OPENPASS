<script lang="ts">
	import {
		ArrowRight,
		CalendarDays,
		CirclePlus,
		Clock3,
		Sparkles,
		Users,
	} from "@lucide/svelte";
	import type { PageData } from "./$types";
	import type { OrganizerEvent } from "$lib/api/analytics";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import StatCard from "$lib/components/dashboard/StatCard.svelte";
	import DashboardSidebar from
		"$lib/components/dashboard/DashboardSidebar.svelte";
	import AttendeeTable from "$lib/components/dashboard/AttendeeTable.svelte";

	let { data } = $props<{ data: PageData }>();

	let events = $derived((data.events ?? []) as OrganizerEvent[]);
	let totalEvents = $derived(events.length);
	let activeEvents = $derived(
		events.filter((event: OrganizerEvent) => event.status === "active").length
	);
	let draftEvents = $derived(
		events.filter((event: OrganizerEvent) => event.status === "draft").length
	);
	let publicEvents = $derived(
		events.filter((event: OrganizerEvent) => event.is_public).length
	);
	let featuredEvent = $derived(events[0]);

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);

		if (Number.isNaN(date.getTime())) {
			return "Date TBD";
		}

		return date.toLocaleDateString("en-IN", {
			weekday: "short",
			day: "numeric",
			month: "short",
			year: "numeric",
		});
	}

	function statusColor(status: string): string {
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
</script>

<div class="space-y-6">
	<div class="border border-[#141414]/10 bg-card p-6 sm:p-7">
		<div
			class="flex flex-col gap-4 sm:flex-row
				sm:items-start sm:justify-between"
		>
			<div class="max-w-3xl">
				<Badge
					variant="secondary"
					class="rounded-none border border-[#141414]/10
						bg-[#141414]/3 px-3 py-1 text-[#3D6B8C]"
				>
					Organizer Home
				</Badge>

				<h1
					class="mt-4 font-serif text-4xl leading-tight
						tracking-tight text-[#141414] sm:text-5xl"
				>
					Plan faster and ship your next event
				</h1>

				<p
					class="mt-3 font-sans text-lg text-[#141414]/68
						sm:max-w-2xl"
				>
					Use this home view to monitor your event pipeline and jump into
					event operations quickly.
				</p>
			</div>

			<Button
				href="/organizer/dashboard/events/new"
				class="h-12 shrink-0 rounded-none border
					border-[#3D6B8C] bg-card px-5 font-sans
					text-sm font-semibold uppercase tracking-[0.12em]
					text-[#3D6B8C] hover:bg-[#3D6B8C]/6"
			>
				<CirclePlus class="size-4" />
				Create Event
			</Button>
		</div>
	</div>

	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
		<StatCard
			label="Total Events"
			value={totalEvents.toLocaleString("en-IN")}
			subtitle="Events in your workspace"
		>
			{#snippet icon()}
				<CalendarDays class="size-4" />
			{/snippet}
		</StatCard>

		<StatCard
			label="Active"
			value={activeEvents.toLocaleString("en-IN")}
			subtitle="Currently running"
		>
			{#snippet icon()}
				<Sparkles class="size-4" />
			{/snippet}
		</StatCard>

		<StatCard
			label="Drafts"
			value={draftEvents.toLocaleString("en-IN")}
			subtitle="Still being prepared"
		>
			{#snippet icon()}
				<Clock3 class="size-4" />
			{/snippet}
		</StatCard>

		<StatCard
			label="Public"
			value={publicEvents.toLocaleString("en-IN")}
			subtitle="Visible to attendees"
		>
			{#snippet icon()}
				<Users class="size-4" />
			{/snippet}
		</StatCard>
	</div>

	{#if events.length > 0}
		<div class="grid grid-cols-1 gap-4 xl:grid-cols-[2fr_1fr]">
			<div class="border border-[#141414]/10 bg-card p-6">
				<div
					class="flex flex-col gap-3 border-b border-[#141414]/8
						pb-5 sm:flex-row sm:items-center sm:justify-between"
				>
					<div>
						<p
							class="font-sans text-xs font-semibold uppercase
								tracking-[0.2em] text-[#3D6B8C]"
						>
							Recent Events
						</p>
						<p class="mt-2 font-sans text-sm text-[#141414]/60">
							Open any event to view analytics, exports, and check-in
							activity.
						</p>
					</div>

					<a
						href="/organizer/dashboard/events/new"
						class="shrink-0 border border-[#141414]/12
							bg-[#141414]/3 px-4 py-2.5 font-sans text-xs
							font-semibold uppercase tracking-[0.15em]
							text-[#141414]/70 transition-colors
							hover:border-[#3D6B8C] hover:text-[#3D6B8C]"
					>
						Add Event
					</a>
				</div>

				<div class="mt-5 space-y-3">
					{#each events.slice(0, 5) as event (event.id)}
						<a
							href="/organizer/dashboard/{event.id}"
							class="block border border-[#141414]/10 bg-[#141414]/2
								p-4 transition-colors hover:border-[#3D6B8C]/30"
						>
							<div class="flex flex-wrap items-center gap-2">
								<p
									class="font-sans text-sm font-medium text-[#141414]"
								>
									{event.title}
								</p>
								<Badge
									variant="secondary"
									class="rounded-none px-2 py-0 text-[10px]
										font-semibold uppercase tracking-[0.1em]
										{statusColor(event.status)}"
								>
									{event.status}
								</Badge>
							</div>

							<p class="mt-2 font-sans text-sm text-[#141414]/62">
								{event.venue}
								<span class="mx-2 text-[#141414]/25">•</span>
								{formatDate(event.start_date)}
							</p>
						</a>
					{/each}
				</div>
			</div>

			<div class="space-y-4">
				<div class="overflow-hidden border border-[#141414]/10 lg:hidden">
					<DashboardSidebar
						events={events}
						activeEventId=""
					/>
				</div>

				{#if featuredEvent}
					<AttendeeTable
						eventId={featuredEvent.id}
						eventTitle={featuredEvent.title}
					/>
				{/if}
			</div>
		</div>
	{:else}
		<div class="border border-[#141414]/10 bg-card p-7">
			<p
				class="font-sans text-xs font-semibold uppercase
					tracking-[0.2em] text-[#3D6B8C]"
			>
				No events yet
			</p>
			<h2 class="mt-4 font-serif text-4xl tracking-tight text-[#141414]">
				Create your first event draft
			</h2>
			<p class="mt-3 max-w-2xl font-sans text-lg text-[#141414]/68">
				Start with a draft event, add registration fields, and publish when
				you are ready.
			</p>

			<div class="mt-6 flex flex-wrap gap-3">
				<Button
					href="/organizer/dashboard/events/new"
					class="h-12 rounded-none border border-[#3D6B8C]
						bg-card px-5 font-sans text-sm font-semibold
						uppercase tracking-[0.12em] text-[#3D6B8C]
						hover:bg-[#3D6B8C]/6"
				>
					Create Event
					<ArrowRight class="size-4" />
				</Button>
			</div>
		</div>
	{/if}
</div>
