<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import { Download } from "@lucide/svelte";
	import { downloadAttendeeExport } from "$lib/api/analytics";

	interface Props {
		eventId: string;
		eventTitle: string;
	}

	let { eventId, eventTitle }: Props = $props();
	let isExporting = $state(false);
	let exportError = $state("");

	function buildFallbackFileName(title: string): string {
		const slug = title
			.trim()
			.toLowerCase()
			.replace(/[^a-z0-9]+/g, "-")
			.replace(/^-+|-+$/g, "");

		if (!slug) {
			return "attendees.csv";
		}

		return `attendees-${slug}.csv`;
	}

	async function handleExport(): Promise<void> {
		if (isExporting) {
			return;
		}

		isExporting = true;
		exportError = "";

		try {
			const { blob, fileName } =
				await downloadAttendeeExport(
					fetch,
					eventId,
					buildFallbackFileName(eventTitle)
				);
			const objectUrl = URL.createObjectURL(blob);
			const link = document.createElement("a");

			link.href = objectUrl;
			link.download = fileName;
			document.body.append(link);
			link.click();
			link.remove();

			window.setTimeout(() => {
				URL.revokeObjectURL(objectUrl);
			}, 0);
		} catch (error) {
			exportError = error instanceof Error
				? error.message
				: "Failed to export attendee data.";
		} finally {
			isExporting = false;
		}
	}
</script>

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
				Attendee Data
			</p>
			<p class="mt-2 font-sans text-sm text-[#141414]/60">
				Export all registrations for
				<span class="font-medium text-[#141414]">
					{eventTitle}
				</span>.
			</p>
		</div>

		<Button
			onclick={handleExport}
			disabled={isExporting}
			variant="outline"
			class="h-11 shrink-0 gap-2 rounded-none
				border-[#141414]/14 px-5 font-sans
				text-sm font-semibold uppercase
				tracking-[0.15em] text-[#141414]
				hover:border-[#3D6B8C]
				hover:text-[#3D6B8C]"
		>
			<Download class="size-4" />
			{isExporting ? "Exporting..." : "Export CSV"}
		</Button>
	</div>

	{#if exportError}
		<p
			class="mt-4 font-sans text-sm text-red-600"
			aria-live="polite"
		>
			{exportError}
		</p>
	{/if}

	<div class="mt-5 overflow-x-auto">
		<table class="w-full text-left">
			<thead>
				<tr
					class="border-b border-[#141414]/8 font-sans
						text-xs font-semibold uppercase
						tracking-[0.2em] text-[#141414]/50"
				>
					<th class="pb-3 pr-6 font-semibold">Column</th>
					<th class="pb-3 pr-6 font-semibold">
						Description
					</th>
				</tr>
			</thead>
			<tbody
				class="font-sans text-sm leading-8 text-[#141414]/80"
			>
				<tr class="border-b border-[#141414]/5">
					<td class="py-2 pr-6 font-medium text-[#141414]">
						Name
					</td>
					<td class="py-2">Attendee full name</td>
				</tr>
				<tr class="border-b border-[#141414]/5">
					<td class="py-2 pr-6 font-medium text-[#141414]">
						Email
					</td>
					<td class="py-2">Registration email</td>
				</tr>
				<tr class="border-b border-[#141414]/5">
					<td class="py-2 pr-6 font-medium text-[#141414]">
						Status
					</td>
					<td class="py-2">
						registered, paid, checked_in, cancelled
					</td>
				</tr>
				<tr class="border-b border-[#141414]/5">
					<td class="py-2 pr-6 font-medium text-[#141414]">
						Checked In At
					</td>
					<td class="py-2">Check-in timestamp</td>
				</tr>
				<tr>
					<td class="py-2 pr-6 font-medium text-[#141414]">
						Registered At
					</td>
					<td class="py-2">Registration timestamp</td>
				</tr>
			</tbody>
		</table>
	</div>
</div>
