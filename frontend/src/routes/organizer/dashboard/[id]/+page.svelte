<script lang="ts">
	import {
		Ticket,
		UserCheck,
		IndianRupee,
		TrendingUp,
	} from "@lucide/svelte";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import StatCard
		from "$lib/components/dashboard/StatCard.svelte";
	import CapacityGauge
		from "$lib/components/dashboard/CapacityGauge.svelte";
	import AttendeeTable
		from "$lib/components/dashboard/AttendeeTable.svelte";
	import type { PageData } from "./$types";

	let { data } = $props<{ data: PageData }>();

	let event = $derived(data.eventDetail);
	let analytics = $derived(data.analytics);

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

	function formatRevenue(value: number): string {
		if (!Number.isFinite(value) || value <= 0) {
			return "₹0";
		}

		return `₹${value.toLocaleString("en-IN", {
			maximumFractionDigits: 0,
		})}`;
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);

		if (Number.isNaN(date.getTime())) {
			return "";
		}

		return date.toLocaleDateString("en-IN", {
			weekday: "short",
			day: "numeric",
			month: "short",
			year: "numeric",
		});
	}

	function formatCheckInRate(): string {
		if (analytics.tickets_sold === 0) {
			return "0%";
		}

		const rate =
			(analytics.check_in_count / analytics.tickets_sold) *
			100;

		return `${rate.toFixed(1)}%`;
	}
</script>

<div class="space-y-6">
	<!-- Event header -->
	<div
		class="border border-[#141414]/10 bg-card p-6 sm:p-7"
	>
		<div
			class="flex flex-col gap-4 sm:flex-row
				sm:items-start sm:justify-between"
		>
			<div class="max-w-2xl">
				<div class="flex flex-wrap gap-2">
					<Badge
						variant="secondary"
						class="rounded-none border border-[#141414]/10
							bg-[#141414]/3 px-3 py-1 text-[#3D6B8C]"
					>
						Event Analytics
					</Badge>
					<Badge
						variant="secondary"
						class="rounded-none px-3 py-1
							{statusColor(event.status)}"
					>
						{event.status}
					</Badge>
				</div>

				<h1
					class="mt-4 font-serif text-4xl leading-tight
						tracking-tight text-[#141414] sm:text-5xl"
				>
					{event.title}
				</h1>

				<p
					class="mt-3 font-sans text-lg text-[#141414]/68"
				>
					{event.venue}
					<span class="mx-2 text-[#141414]/30">•</span>
					{formatDate(event.start_date)}
				</p>
			</div>

			{#if event.is_public}
				<a
					href="/events/{event.slug}"
					class="shrink-0 border border-[#141414]/12
						bg-[#141414]/3 px-5 py-3 font-sans text-sm
						font-semibold uppercase tracking-[0.15em]
						text-[#141414]/70 transition-colors
						hover:border-[#3D6B8C]
						hover:text-[#3D6B8C]"
				>
					View Public Page
				</a>
			{/if}
		</div>
	</div>

	<!-- KPI row -->
	<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
		<StatCard
			label="Tickets Sold"
			value={analytics.tickets_sold.toLocaleString("en-IN")}
			subtitle="Total registrations"
		>
			{#snippet icon()}
				<Ticket class="size-4" />
			{/snippet}
		</StatCard>

		<StatCard
			label="Checked In"
			value={analytics.check_in_count.toLocaleString("en-IN")}
			subtitle="{formatCheckInRate()} check-in rate"
		>
			{#snippet icon()}
				<UserCheck class="size-4" />
			{/snippet}
		</StatCard>

		<StatCard
			label="Revenue"
			value={formatRevenue(analytics.total_revenue)}
			subtitle="Total collected"
		>
			{#snippet icon()}
				<IndianRupee class="size-4" />
			{/snippet}
		</StatCard>

		<StatCard
			label="Capacity"
			value="{analytics.tickets_sold} / {analytics.capacity}"
			subtitle="{analytics.capacity_pct.toFixed(1)}% filled"
		>
			{#snippet icon()}
				<TrendingUp class="size-4" />
			{/snippet}
		</StatCard>
	</div>

	<!-- Bottom grid: Gauge + Attendee export -->
	<div
		class="grid grid-cols-1 gap-4
			lg:grid-cols-[18rem_1fr]"
	>
		<CapacityGauge
			percentage={analytics.capacity_pct}
			ticketsSold={analytics.tickets_sold}
			capacity={analytics.capacity}
		/>

		<AttendeeTable
			eventId={event.id}
			eventTitle={event.title}
		/>
	</div>
</div>
