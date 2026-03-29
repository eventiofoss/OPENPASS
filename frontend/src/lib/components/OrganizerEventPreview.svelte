<script lang="ts">
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import type { FormFieldType } from "$lib/api/events";

	type VisibilityMode = "public" | "private";
	type TicketMode = "free" | "paid";

	interface PreviewField {
		id: string;
		label: string;
		type: FormFieldType;
		required: boolean;
		options: string[];
	}

	interface Props {
		title: string;
		description: string;
		venue: string;
		startDate: string;
		capacity: number;
		visibility: VisibilityMode;
		ticketMode: TicketMode;
		price: number;
		fields: PreviewField[];
	}

	let {
		title,
		description,
		venue,
		startDate,
		capacity,
		visibility,
		ticketMode,
		price,
		fields
	}: Props = $props();

	const formInputClass =
		"h-12 rounded-none border-[#141414]/14 bg-card px-4 " +
		"font-sans text-base text-[#141414] placeholder:text-[#141414]/40";

	function formatStartDate(value: string): string {
		if (!value) {
			return "Choose a start date and time";
		}

		const date = new Date(value);

		if (Number.isNaN(date.getTime())) {
			return "Choose a start date and time";
		}

		return date.toLocaleString("en-IN", {
			weekday: "short",
			day: "numeric",
			month: "short",
			hour: "2-digit",
			minute: "2-digit",
			hour12: true,
		});
	}

	function formatPrice(value: number): string {
		if (!Number.isFinite(value) || value <= 0) {
			return "Free";
		}

		return `₹${value.toLocaleString("en-IN", {
			maximumFractionDigits: 2,
		})}`;
	}

	function placeholderFor(type: FormFieldType): string {
		switch (type) {
			case "email":
				return "attendee@email.com";
			case "number":
				return "Enter a number";
			default:
				return "Type your answer";
		}
	}
</script>

<section class="space-y-6 lg:sticky lg:top-6">
	<div class="overflow-hidden border border-[#141414]/14 bg-card">
		<div
			class="bg-[linear-gradient(135deg,#3D6B8C_0%,#22445B_55%,#141414_100%)]
				p-7 text-white sm:p-8"
		>
			<div class="flex flex-wrap gap-2">
				<Badge
					variant="secondary"
					class="rounded-none bg-white/12 px-3 py-1 text-white"
				>
					{visibility === "public" ? "Public event" : "Private event"}
				</Badge>
				<Badge
					variant="secondary"
					class="rounded-none bg-white/12 px-3 py-1 text-white"
				>
					{ticketMode === "free" ? "Free entry" : "Paid ticket"}
				</Badge>
			</div>

			<h2
				class="mt-5 font-serif text-4xl leading-tight tracking-tight
					sm:text-5xl"
			>
				{title.trim() || "Your event title will show here"}
			</h2>

			<p
				class="mt-4 max-w-2xl font-sans text-lg leading-8 text-white/78
					sm:text-xl"
			>
				{venue.trim() || "Venue details"}
				<span class="mx-2 text-white/35">•</span>
				{formatStartDate(startDate)}
			</p>
		</div>

		<div class="grid gap-4 p-7 sm:grid-cols-2 sm:p-8">
			<div class="border border-[#141414]/10 bg-[#141414]/3 p-4">
				<p
					class="font-sans text-xs font-semibold uppercase
						tracking-[0.2em] text-[#3D6B8C]"
				>
					Price
				</p>
				<p class="mt-2 font-serif text-3xl text-[#141414]">
					{ticketMode === "free" ? "Free" : formatPrice(price)}
				</p>
			</div>

			<div class="border border-[#141414]/10 bg-[#141414]/3 p-4">
				<p
					class="font-sans text-xs font-semibold uppercase
						tracking-[0.2em] text-[#3D6B8C]"
				>
					Capacity
				</p>
				<p class="mt-2 font-serif text-3xl text-[#141414]">
					{capacity}
				</p>
			</div>
		</div>

		<div class="border-t border-[#141414]/10 p-7 sm:p-8">
			<p
				class="font-sans text-sm font-semibold uppercase
					tracking-[0.3em] text-[#3D6B8C]"
			>
				What attendees will read
			</p>
			<p
				class="mt-4 whitespace-pre-line font-sans text-lg leading-8
					text-[#141414]/80 sm:text-xl"
			>
				{description.trim() ||
					"Add a description to introduce the vibe, agenda, " +
						"and reason to register."}
			</p>
		</div>
	</div>

	<div class="border border-[#141414]/14 bg-card p-7 sm:p-8">
		<div
			class="flex flex-col gap-3 border-b border-[#141414]/10 pb-5
				sm:flex-row sm:items-end sm:justify-between"
		>
			<div>
				<p
					class="font-sans text-sm font-semibold uppercase
						tracking-[0.3em] text-[#3D6B8C]"
				>
					Registration Preview
				</p>
				<p class="mt-2 font-sans text-base text-[#141414]/68">
					Name and email are always collected first.
				</p>
			</div>
			<p class="font-sans text-base text-[#141414]/68">
				{fields.length} custom question{fields.length === 1 ? "" : "s"}
			</p>
		</div>

		<div class="mt-6 space-y-4">
			<div class="space-y-2">
				<label
					class="font-sans text-lg font-medium text-[#141414]"
					for="preview-name"
				>
					Full name
				</label>
				<Input
					id="preview-name"
					disabled
					placeholder="Attendee name"
					class={formInputClass}
				/>
			</div>

			<div class="space-y-2">
				<label
					class="font-sans text-lg font-medium text-[#141414]"
					for="preview-email"
				>
					Email address
				</label>
				<Input
					id="preview-email"
					disabled
					type="email"
					placeholder="you@email.com"
					class={formInputClass}
				/>
			</div>

			{#each fields as field (field.id)}
				<div class="space-y-2">
					<div class="flex items-center justify-between gap-3">
						<label
							class="font-sans text-lg font-medium text-[#141414]"
						>
							{field.label}
						</label>

						{#if field.required}
							<span class="font-sans text-sm text-[#3D6B8C]">
								Required
							</span>
						{/if}
					</div>

					{#if field.type === "checkbox"}
						<label
							class="flex items-center gap-3 border border-[#141414]/14
								bg-[#141414]/3 px-4 py-3 font-sans text-base
								text-[#141414]/82"
						>
							<span class="h-5 w-5 border border-[#141414]/18"></span>
							{field.label}
						</label>
					{:else if field.type === "select"}
						<select
							disabled
							class="h-12 w-full rounded-none border border-[#141414]/14
								bg-card px-4 font-sans text-base text-[#141414]/68"
						>
							<option>
								{field.options[0] || "Choose an option"}
							</option>
						</select>
					{:else}
						<Input
							disabled
							type={field.type === "number"
								? "number"
								: field.type === "email"
									? "email"
									: "text"}
							placeholder={placeholderFor(field.type)}
							class={formInputClass}
						/>
					{/if}
				</div>
			{/each}
		</div>

		<div
			class="mt-6 rounded-none bg-[#141414] px-5 py-4 text-center
				font-sans text-lg text-white"
		>
			{ticketMode === "free"
				? "Register for Free"
				: `Continue to Payment • ${formatPrice(price)}`}
		</div>

		<p class="mt-3 font-sans text-sm leading-6 text-[#141414]/62">
			{ticketMode === "free"
				? "Attendees finish registration as soon as this form is " +
					"submitted."
				: "The payment button appears after the attendee completes " +
					"these fields."}
		</p>

		{#if visibility === "private"}
			<div
				class="mt-5 border border-[#141414]/12 bg-[#141414]/3 p-4
					font-sans text-base leading-7 text-[#141414]/74"
			>
				Private events stay behind a secret access token. Only invitees
				with the private link can open the event page.
			</div>
		{/if}
	</div>
</section>
