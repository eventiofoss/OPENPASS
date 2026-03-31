<script lang="ts">
	import { goto, invalidate } from "$app/navigation";
	import {
		CalendarDays,
		CircleDollarSign,
		FileText,
		LockKeyhole,
		MapPinned,
		Plus,
		Ticket,
		Trash2,
		Users,
	} from "@lucide/svelte";
	import OrganizerEventPreview from
		"$lib/components/OrganizerEventPreview.svelte";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { apiFetch } from "$lib/api/http";
	import {
		organizerAnalyticsDependency,
		organizerEventDependency,
		organizerEventsDependency,
	} from "$lib/utils/organizer-dashboard";
	import {
		createOrganizerEvent,
		setOrganizerEventFormFields,
		type CreateOrganizerEventInput,
		type FormFieldType,
		type OrganizerEventRecord,
		type OrganizerFormFieldInput,
	} from "$lib/api/events";

	type VisibilityMode = "public" | "private";
	type TicketMode = "free" | "paid";

	interface DraftFormField {
		id: string;
		label: string;
		type: FormFieldType;
		required: boolean;
		optionsText: string;
	}

	interface CreatedEventState {
		id: string;
		slug: string;
		fieldCount: number;
		formSaved: boolean;
	}

	const fieldTypeOptions: Array<{
		value: FormFieldType;
		label: string;
		description: string;
	}> = [
		{
			value: "text",
			label: "Short text",
			description: "Names, company, city, or any simple answer.",
		},
		{
			value: "email",
			label: "Email",
			description: "Extra contact email or alternate inbox.",
		},
		{
			value: "select",
			label: "Select",
			description: "Meal choices, ticket waves, or shirt sizes.",
		},
		{
			value: "checkbox",
			label: "Checkbox",
			description: "Consent, waivers, or yes-or-no confirmations.",
		},
		{
			value: "number",
			label: "Number",
			description: "Age, group size, or numeric preferences.",
		},
	];

	const selectClass =
		"h-12 w-full rounded-none border border-[#141414]/14 bg-card " +
		"px-4 font-sans text-base text-[#141414] outline-none " +
		"focus:border-[#3D6B8C]";

	const textareaClass =
		"min-h-36 w-full rounded-none border border-[#141414]/14 bg-card " +
		"px-4 py-4 font-sans text-base leading-7 text-[#141414] " +
		"outline-none placeholder:text-[#141414]/45 focus:border-[#3D6B8C]";

	const eventInputClass =
		"h-14 rounded-none border-[#141414]/14 bg-card px-4 font-sans " +
		"text-base text-[#141414] placeholder:text-[#141414]/45";

	let eventDetails = $state(createEmptyEventDetails());
	let visibilityMode = $state<VisibilityMode>("public");
	let ticketMode = $state<TicketMode>("free");
	let ticketPrice = $state(499);
	let attendeeFields = $state<DraftFormField[]>([]);
	let isSubmitting = $state(false);
	let isRetryingFields = $state(false);
	let isPublishing = $state(false);
	let submitError = $state("");
	let publishError = $state<string | null>(null);
	let createdEvent = $state<CreatedEventState | null>(null);
	let currentEventId = $derived(createdEvent?.id ?? "");

	function createEmptyEventDetails() {
		return {
			title: "",
			description: "",
			venue: "",
			startDate: buildDefaultStartDate(),
			capacity: 120,
		};
	}

	function buildDefaultStartDate(): string {
		const nextDay = new Date();

		nextDay.setDate(nextDay.getDate() + 7);
		nextDay.setHours(18, 30, 0, 0);

		return toDateTimeLocal(nextDay);
	}

	function toDateTimeLocal(date: Date): string {
		const year = date.getFullYear();
		const month = String(date.getMonth() + 1).padStart(2, "0");
		const day = String(date.getDate()).padStart(2, "0");
		const hours = String(date.getHours()).padStart(2, "0");
		const minutes = String(date.getMinutes()).padStart(2, "0");

		return `${year}-${month}-${day}T${hours}:${minutes}`;
	}

	function createFieldId(): string {
		if (
			typeof crypto !== "undefined" &&
			typeof crypto.randomUUID === "function"
		) {
			return crypto.randomUUID();
		}

		return `field-${Date.now()}-${Math.random()
			.toString(36)
			.slice(2, 8)}`;
	}

	function addField(type: FormFieldType = "text") {
		attendeeFields = [
			...attendeeFields,
			{
				id: createFieldId(),
				label: "",
				type,
				required: false,
				optionsText: "",
			},
		];
	}

	function removeField(fieldId: string) {
		attendeeFields = attendeeFields.filter(
			(field) => field.id !== fieldId
		);
	}

	function parseOptions(optionsText: string): string[] {
		return optionsText
			.split(/\n|,/)
			.map((entry) => entry.trim())
			.filter(Boolean);
	}

	function previewFields() {
		return attendeeFields
			.filter((field) => field.label.trim() !== "")
			.map((field) => ({
				id: field.id,
				label: field.label.trim(),
				type: field.type,
				required: field.required,
				options: parseOptions(field.optionsText),
			}));
	}

	function buildFieldName(
		label: string,
		index: number,
		usedNames: Set<string>
	): string {
		const base = label
			.toLowerCase()
			.trim()
			.replace(/[^a-z0-9]+/g, "_")
			.replace(/^_+|_+$/g, "")
			.slice(0, 40);

		let name = base || `field_${index + 1}`;
		let suffix = 2;

		while (usedNames.has(name)) {
			name = `${base || `field_${index + 1}`}_${suffix}`;
			suffix += 1;
		}

		usedNames.add(name);
		return name;
	}

	function buildFormFieldPayload(): OrganizerFormFieldInput[] {
		const usedNames = new Set<string>();

		return attendeeFields
			.filter((field) => field.label.trim() !== "")
			.map((field, index) => {
				const input: OrganizerFormFieldInput = {
					name: buildFieldName(
						field.label,
						index,
						usedNames
					),
					type: field.type,
					label: field.label.trim(),
					required: field.required,
				};

				if (field.type === "select") {
					input.options = parseOptions(field.optionsText);
				}

				return input;
			});
	}

	function segmentClass(active: boolean): string {
		return active
			? "border-[#3D6B8C] bg-[#3D6B8C]/10 text-[#141414]"
			: "border-[#141414]/12 bg-card text-[#141414]/68 " +
				"hover:border-[#141414]/24 hover:bg-[#141414]/3";
	}

	function toFiniteNumber(value: unknown): number {
		const numberValue = Number(value);

		if (!Number.isFinite(numberValue)) {
			return Number.NaN;
		}

		return numberValue;
	}

	function formatPriceSummary(value: number): string {
		if (!Number.isFinite(value) || value <= 0) {
			return "Free";
		}

		return `₹${value.toLocaleString("en-IN", {
			maximumFractionDigits: 2,
		})}`;
	}

	function getErrorMessage(error: unknown): string {
		if (error instanceof Error && error.message.trim() !== "") {
			return error.message;
		}

		return "Something went wrong while creating the event.";
	}

	function validateDraft(): string | null {
		if (eventDetails.title.trim() === "") {
			return "Add an event title before creating the draft.";
		}

		if (eventDetails.description.trim() === "") {
			return "Add a description so the event page has real content.";
		}

		if (eventDetails.venue.trim() === "") {
			return "Add a venue for the event page.";
		}

		const startDate = new Date(eventDetails.startDate);
		const capacityValue = toFiniteNumber(eventDetails.capacity);
		const priceValue = toFiniteNumber(ticketPrice);

		if (Number.isNaN(startDate.getTime())) {
			return "Choose a valid start date and time.";
		}

		if (!Number.isFinite(capacityValue) || capacityValue < 0) {
			return "Capacity must be zero or a positive number.";
		}

		if (
			ticketMode === "paid" &&
			(!Number.isFinite(priceValue) || priceValue <= 0)
		) {
			return "Paid events need a ticket price above zero.";
		}

		for (const field of attendeeFields) {
			if (field.label.trim() === "") {
				return "Each custom question needs a label or should be removed.";
			}

			if (
				field.type === "select" &&
				parseOptions(field.optionsText).length === 0
			) {
				return `Add at least one option for "${field.label}".`;
			}
		}

		return null;
	}

	async function saveCustomFields(eventId: string): Promise<number> {
		const fields = buildFormFieldPayload();

		if (fields.length === 0) {
			return 0;
		}

		await setOrganizerEventFormFields(fetch, eventId, fields);
		return fields.length;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		submitError = "";
		createdEvent = null;

		const validationError = validateDraft();

		if (validationError) {
			submitError = validationError;
			return;
		}

		isSubmitting = true;

		let draftEvent: OrganizerEventRecord | null = null;

		try {
			const capacityValue = toFiniteNumber(eventDetails.capacity);
			const priceValue = toFiniteNumber(ticketPrice);
			const payload: CreateOrganizerEventInput = {
				title: eventDetails.title.trim(),
				description: eventDetails.description.trim(),
				start_date: new Date(eventDetails.startDate).toISOString(),
				venue: eventDetails.venue.trim(),
				capacity: capacityValue,
				price: ticketMode === "paid" ? priceValue : 0,
				is_public: visibilityMode === "public",
			};

			draftEvent = await createOrganizerEvent(fetch, payload);

			const fieldCount = await saveCustomFields(draftEvent.id);

			createdEvent = {
				id: draftEvent.id,
				slug: draftEvent.slug,
				fieldCount,
				formSaved: true,
			};
			await invalidate(organizerEventsDependency);
		} catch (error) {
			if (draftEvent) {
				await invalidate(organizerEventsDependency);
				createdEvent = {
					id: draftEvent.id,
					slug: draftEvent.slug,
					fieldCount: 0,
					formSaved: false,
				};
				submitError =
					getErrorMessage(error) +
					" The draft event exists, but the attendee " +
					"form still needs to be saved.";
			} else {
				submitError = getErrorMessage(error);
			}
		} finally {
			isSubmitting = false;
		}
	}

	async function retryFormSave() {
		if (!createdEvent || createdEvent.formSaved) {
			return;
		}

		submitError = "";
		isRetryingFields = true;

		try {
			const fieldCount = await saveCustomFields(createdEvent.id);

			createdEvent = {
				...createdEvent,
				fieldCount,
				formSaved: true,
			};
		} catch (error) {
			submitError = getErrorMessage(error);
		} finally {
			isRetryingFields = false;
		}
	}

	async function publishEvent() {
		if (!currentEventId) {
			return;
		}

		isPublishing = true;
		publishError = null;

		try {
			const response = await apiFetch(fetch, `/api/events/${currentEventId}`, {
				method: "PATCH",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify({
					status: "active",
				}),
			});

			if (!response.ok) {
				throw new Error("Failed to activate event. Please try again.");
			}

			await response.json();
			await Promise.all([
				invalidate(organizerEventsDependency),
				invalidate(
					organizerEventDependency(currentEventId)
				),
				invalidate(
					organizerAnalyticsDependency(
						currentEventId
					)
				),
			]);
			await goto(`/organizer/dashboard/${currentEventId}`);
		} catch (err: unknown) {
			publishError = err instanceof Error
				? err.message
				: "Failed to activate event. Please try again.";
		} finally {
			isPublishing = false;
		}
	}

	function resetForm() {
		eventDetails = createEmptyEventDetails();
		visibilityMode = "public";
		ticketMode = "free";
		ticketPrice = 499;
		attendeeFields = [];
		isSubmitting = false;
		isRetryingFields = false;
		submitError = "";
		publishError = null;
		createdEvent = null;
	}
</script>

<svelte:head>
	<title>Create Event | Open Pass</title>
	<meta
		name="description"
		content="Build a new event draft with attendee registration fields,
		ticket settings, and a live public-page preview."
	/>
</svelte:head>

<section>
	<div
		class="border border-[#141414]/14 bg-card p-6 text-[#141414]
			sm:p-8"
	>
				<div class="border-b border-[#141414]/10 pb-7">
					<div class="max-w-3xl">
						<div class="flex flex-wrap gap-2">
							<Badge
								variant="secondary"
								class="rounded-none border border-[#141414]/10
									bg-[#141414]/3 px-3 py-1 text-[#3D6B8C]"
							>
								Organizer Workspace
							</Badge>
							<Badge
								variant="secondary"
								class="rounded-none border border-[#141414]/10
									bg-[#141414]/3 px-3 py-1 text-[#3D6B8C]"
							>
								Event Draft Builder
							</Badge>
						</div>

						<h1
							class="mt-5 font-serif text-5xl leading-tight
								tracking-tight sm:text-6xl lg:text-7xl"
						>
							Create an event page and the registration form in one
							flow
						</h1>

						<p
							class="mt-5 max-w-2xl font-sans text-lg leading-8
								text-[#141414]/78 sm:text-xl"
						>
							Set the event details attendees should see, choose
							whether it is free or paid, and build the extra
							questions they need to answer before registering.
						</p>
					</div>
				</div>

				<div class="mt-8 grid gap-8 lg:grid-cols-[minmax(0,1.25fr)_26rem]">
					<form class="space-y-6" onsubmit={handleSubmit}>
						<div class="border border-[#141414]/14 bg-card p-6 sm:p-8">
							<div class="flex items-start gap-4">
								<div
									class="flex h-11 w-11 shrink-0 items-center
										justify-center border border-[#141414]/12
										bg-[#141414]/3 text-[#3D6B8C]"
								>
									<FileText class="size-5" />
								</div>

								<div>
									<p
										class="font-sans text-sm font-semibold
											uppercase tracking-[0.3em]
											text-[#3D6B8C]"
									>
										Event Details
									</p>
									<h2 class="mt-3 font-serif text-4xl">
										What should appear on the event page?
									</h2>
									<p
										class="mt-3 max-w-2xl font-sans text-base
											leading-7 text-[#141414]/70"
									>
										These fields power the public-facing event
										page: title, story, venue, time, capacity,
										and ticketing state.
									</p>
								</div>
							</div>

							<div class="mt-8 space-y-5">
								<div class="space-y-2">
									<label
										for="event-title"
										class="font-sans text-lg font-medium
											text-[#141414]"
									>
										Event title
									</label>
									<Input
										id="event-title"
										name="title"
										bind:value={eventDetails.title}
										placeholder="Sunrise founder breakfast"
										class={eventInputClass}
									/>
								</div>

								<div class="space-y-2">
									<label
										for="event-description"
										class="font-sans text-lg font-medium
											text-[#141414]"
									>
										Description
									</label>
									<textarea
										id="event-description"
										name="description"
										bind:value={eventDetails.description}
										placeholder="Describe the experience, the agenda, and why someone should register."
										class={textareaClass}
									></textarea>
								</div>

								<div class="grid gap-5 sm:grid-cols-2">
									<div class="space-y-2">
										<label
											for="event-date"
											class="font-sans text-lg font-medium
												text-[#141414]"
										>
											Start date and time
										</label>
										<div class="relative">
											<CalendarDays
												class="pointer-events-none absolute
													left-4 top-1/2 size-5
													-translate-y-1/2
													text-[#3D6B8C]"
											/>
											<Input
												id="event-date"
												name="start_date"
												type="datetime-local"
												bind:value={eventDetails.startDate}
												class="h-14 rounded-none
													border-[#141414]/14 bg-card
													pl-12 pr-4 font-sans text-lg
													text-[#141414]"
											/>
										</div>
									</div>

									<div class="space-y-2">
										<label
											for="event-capacity"
											class="font-sans text-lg font-medium
												text-[#141414]"
										>
											Capacity
										</label>
										<div class="relative">
											<Users
												class="pointer-events-none absolute
													left-4 top-1/2 size-5
													-translate-y-1/2
													text-[#3D6B8C]"
											/>
											<Input
												id="event-capacity"
												name="capacity"
												type="number"
												min="0"
												bind:value={eventDetails.capacity}
												class="h-14 rounded-none
													border-[#141414]/14 bg-card
													pl-12 pr-4 font-sans text-lg
													text-[#141414]"
											/>
										</div>
									</div>
								</div>

								<div class="space-y-2">
									<label
										for="event-venue"
										class="font-sans text-lg font-medium
											text-[#141414]"
									>
										Venue
									</label>
									<div class="relative">
										<MapPinned
											class="pointer-events-none absolute left-4
												top-1/2 size-5 -translate-y-1/2
												text-[#3D6B8C]"
										/>
										<Input
											id="event-venue"
											name="venue"
											bind:value={eventDetails.venue}
											placeholder="The Glass House, Bengaluru"
											class="h-14 rounded-none
												border-[#141414]/14 bg-card pl-12
												pr-4 font-sans text-lg
												text-[#141414]
												placeholder:text-[#141414]/45"
										/>
									</div>
								</div>
							</div>
						</div>

						<div class="border border-[#141414]/14 bg-card p-6 sm:p-8">
							<div class="flex items-start gap-4">
								<div
									class="flex h-11 w-11 shrink-0 items-center
										justify-center border border-[#141414]/12
										bg-[#141414]/3 text-[#3D6B8C]"
								>
									<CircleDollarSign class="size-5" />
								</div>

								<div>
									<p
										class="font-sans text-sm font-semibold
											uppercase tracking-[0.3em]
											text-[#3D6B8C]"
									>
										Access and Ticketing
									</p>
									<h2 class="mt-3 font-serif text-4xl">
										Public or private, free or paid
									</h2>
									<p
										class="mt-3 max-w-2xl font-sans text-base
											leading-7 text-[#141414]/70"
									>
										Use the toggles below to decide whether the
										event needs a secret invite link and whether
										attendees should continue to a payment step.
									</p>
								</div>
							</div>

							<div class="mt-8 grid gap-6 lg:grid-cols-2">
								<div>
									<p
										class="font-sans text-base font-semibold
											uppercase tracking-[0.2em]
											text-[#3D6B8C]"
									>
										Visibility
									</p>

									<div class="mt-4 grid gap-3">
										<button
											type="button"
											class={`border p-4 text-left transition-colors ${segmentClass(
												visibilityMode === "public"
											)}`}
											onclick={() => {
												visibilityMode = "public";
											}}
										>
											<p class="font-sans text-lg font-medium">
												Public event
											</p>
											<p
												class="mt-2 font-sans text-sm leading-6
													text-[#141414]/68"
											>
												The public event page can be listed and
												shared once the event becomes active.
											</p>
										</button>

										<button
											type="button"
											class={`border p-4 text-left transition-colors ${segmentClass(
												visibilityMode === "private"
											)}`}
											onclick={() => {
												visibilityMode = "private";
											}}
										>
											<div class="flex items-center gap-2">
												<LockKeyhole class="size-4" />
												<p class="font-sans text-lg font-medium">
													Private event
												</p>
											</div>
											<p
												class="mt-2 font-sans text-sm leading-6
													text-[#141414]/68"
											>
												Attendees will need the backend-generated
												secret link token to open the page.
											</p>
										</button>
									</div>
								</div>

								<div>
									<p
										class="font-sans text-base font-semibold
											uppercase tracking-[0.2em]
											text-[#3D6B8C]"
									>
										Ticket type
									</p>

									<div class="mt-4 grid gap-3">
										<button
											type="button"
											class={`border p-4 text-left transition-colors ${segmentClass(
												ticketMode === "free"
											)}`}
											onclick={() => {
												ticketMode = "free";
											}}
										>
											<p class="font-sans text-lg font-medium">
												Free registration
											</p>
											<p
												class="mt-2 font-sans text-sm leading-6
													text-[#141414]/68"
											>
												Attendees submit the form and receive
												their pass immediately.
											</p>
										</button>

										<button
											type="button"
											class={`border p-4 text-left transition-colors ${segmentClass(
												ticketMode === "paid"
											)}`}
											onclick={() => {
												ticketMode = "paid";
											}}
										>
											<p class="font-sans text-lg font-medium">
												Paid registration
											</p>
											<p
												class="mt-2 font-sans text-sm leading-6
													text-[#141414]/68"
											>
												Show a payment continuation after the
												registration form is filled out.
											</p>
										</button>
									</div>

									{#if ticketMode === "paid"}
										<div class="mt-5 space-y-2">
											<label
												for="ticket-price"
												class="font-sans text-lg font-medium
													text-[#141414]"
											>
												Ticket price
											</label>
											<div class="relative">
												<Ticket
													class="pointer-events-none absolute
														left-4 top-1/2 size-5
														-translate-y-1/2
														text-[#3D6B8C]"
												/>
												<Input
													id="ticket-price"
													type="number"
													min="0"
													step="0.01"
													bind:value={ticketPrice}
													class="h-14 rounded-none
														border-[#141414]/14 bg-card
														pl-12 pr-4 font-sans text-lg
														text-[#141414]"
												/>
											</div>
										</div>
									{:else}
										<div
											class="mt-5 border border-[#141414]/10
												bg-[#141414]/3 p-4 font-sans text-base
												leading-7 text-[#141414]/74"
										>
											The public page will show
											<span class="font-semibold text-[#141414]">
												Free
											</span>
											and the registration button will not lead
											into payment.
										</div>
									{/if}
								</div>

							</div>

							<div
								class="mt-6 border border-[#141414]/10 bg-[#141414]/3
									p-5"
							>
								<p
									class="font-sans text-base font-medium
										text-[#141414]"
								>
									Current attendee flow
								</p>
								<p
									class="mt-2 font-sans text-base leading-7
										text-[#141414]/70"
								>
									{visibilityMode === "public"
										? "Attendees can discover the event once it is active."
										: "Attendees need the private invite link."}
									{" "}
									{ticketMode === "free"
										? "They submit the form and get their pass."
										: `They submit the form and continue to a ${formatPriceSummary(
												ticketPrice
											)} payment step.`}
								</p>
							</div>
						</div>

						<div class="border border-[#141414]/14 bg-card p-6 sm:p-8">
							<div class="flex items-start gap-4">
								<div
									class="flex h-11 w-11 shrink-0 items-center
										justify-center border border-[#141414]/12
										bg-[#141414]/3 text-[#3D6B8C]"
								>
									<Plus class="size-5" />
								</div>

								<div>
									<p
										class="font-sans text-sm font-semibold
											uppercase tracking-[0.3em]
											text-[#3D6B8C]"
									>
										Attendee Form Builder
									</p>
									<h2 class="mt-3 font-serif text-4xl">
										What extra details should attendees fill in?
									</h2>
									<p
										class="mt-3 max-w-2xl font-sans text-base
											leading-7 text-[#141414]/70"
									>
										Full name and email are already part of
										registration. Add any custom questions here,
										and they will be saved to the event form
										schema after the draft is created.
									</p>
								</div>
							</div>

							<div class="mt-7 grid gap-3 md:grid-cols-2">
								{#each fieldTypeOptions as option}
									<Button
										type="button"
										variant="ghost"
										class="h-auto w-full rounded-none border
											border-[#141414]/12 bg-card px-4 py-3
											font-sans text-left text-[#141414]
											whitespace-normal justify-start items-start
											flex-col gap-1.5
											hover:bg-[#141414]/3"
										onclick={() => {
											addField(option.value);
										}}
									>
										<span class="block w-full text-base font-medium">
											Add {option.label}
										</span>
										<span
											class="block w-full text-sm
												leading-6 text-[#141414]/60"
										>
											{option.description}
										</span>
									</Button>
								{/each}
							</div>

							{#if attendeeFields.length === 0}
								<div
									class="mt-6 border border-dashed border-[#141414]/18
										bg-[#141414]/3 p-6"
								>
									<p class="font-sans text-lg text-[#141414]">
										No custom questions yet.
									</p>
									<p
										class="mt-2 max-w-2xl font-sans text-base
											leading-7 text-[#141414]/68"
									>
										Add things like meal preference, community,
										T-shirt size, company, or consent checkboxes.
									</p>
								</div>
							{:else}
								<div class="mt-6 space-y-5">
									{#each attendeeFields as field, index (field.id)}
										<div
											class="border border-[#141414]/12 bg-[#141414]/3
												p-5 sm:p-6"
										>
											<div
												class="flex flex-col gap-3 border-b
													border-[#141414]/10 pb-5 sm:flex-row
													sm:items-start sm:justify-between"
											>
												<div>
													<p
														class="font-sans text-base
															font-semibold uppercase
															tracking-[0.2em]
															text-[#3D6B8C]"
													>
														Question {index + 1}
													</p>
													<p
														class="mt-2 font-sans text-base
															leading-7 text-[#141414]/68"
													>
														Choose the field type, write the
														label attendees will see, and mark
														it required if needed.
													</p>
												</div>

												<Button
													type="button"
													variant="ghost"
													class="h-11 rounded-none border
														border-[#141414]/12 px-4
														font-sans text-[#141414]
														hover:bg-[#141414]/6"
													onclick={() => {
														removeField(field.id);
													}}
												>
													<Trash2 class="size-4" />
													Remove
												</Button>
											</div>

											<div class="mt-5 grid gap-5 sm:grid-cols-2">
												<div class="space-y-2">
													<label
														class="font-sans text-lg font-medium
															text-[#141414]"
														for={`field-label-${field.id}`}
													>
														Field label
													</label>
													<Input
														id={`field-label-${field.id}`}
														bind:value={field.label}
														placeholder="Company or community"
														class={eventInputClass}
													/>
												</div>

												<div class="space-y-2">
													<label
														class="font-sans text-lg font-medium
															text-[#141414]"
														for={`field-type-${field.id}`}
													>
														Field type
													</label>
													<select
														id={`field-type-${field.id}`}
														bind:value={field.type}
														class={selectClass}
													>
														{#each fieldTypeOptions as option}
															<option value={option.value}>
																{option.label}
															</option>
														{/each}
													</select>
												</div>
											</div>

											<div class="mt-5 flex items-center gap-3">
												<input
													id={`field-required-${field.id}`}
													type="checkbox"
													bind:checked={field.required}
													class="h-5 w-5 rounded-none border
														border-[#141414]/18 accent-[#3D6B8C]"
												/>
												<label
													for={`field-required-${field.id}`}
													class="font-sans text-base
														text-[#141414]/76"
												>
													Attendees must answer this question
												</label>
											</div>

											{#if field.type === "select"}
												<div class="mt-5 space-y-2">
													<label
														class="font-sans text-lg font-medium
															text-[#141414]"
														for={`field-options-${field.id}`}
													>
														Options
													</label>
													<textarea
														id={`field-options-${field.id}`}
														bind:value={field.optionsText}
														placeholder={"Breakfast\nLunch\nDinner"}
														class="min-h-28 w-full
															rounded-none border
															border-[#141414]/14 bg-card
															px-4 py-4 font-sans text-base
															leading-7 text-[#141414]
															outline-none
															placeholder:text-[#141414]/45
															focus:border-[#3D6B8C]"
													></textarea>
													<p
														class="font-sans text-sm leading-6
															text-[#141414]/60"
													>
														Add one option per line or separate
														them with commas.
													</p>
												</div>
											{/if}
										</div>
									{/each}
								</div>
							{/if}
						</div>

						<div class="border border-[#141414]/14 bg-card p-6 sm:p-8">
							<p
								class="font-sans text-sm font-semibold uppercase
									tracking-[0.3em] text-[#3D6B8C]"
							>
								Create Draft
							</p>
							<h2 class="mt-3 font-serif text-4xl">
								Save the event and the attendee form
							</h2>

							<div class="mt-6 space-y-4">
								{#if submitError}
									<div
										class="border border-[#a63c3c]/18 bg-[#a63c3c]/8
											p-4 font-sans text-base leading-7
											text-[#7a1f1f]"
									>
										{submitError}
									</div>
								{/if}

								{#if createdEvent}
									<div
										class="border border-[#3D6B8C]/16 bg-[#3D6B8C]/8
											p-4 font-sans text-base leading-7
											text-[#1f4c67]"
									>
										Draft created with slug
										<span class="font-semibold">
											{createdEvent.slug}
										</span>
										and event ID
										<span class="font-semibold">
											{createdEvent.id}
										</span>.
										{#if createdEvent.formSaved}
											{" "}
											{createdEvent.fieldCount} custom question
											{createdEvent.fieldCount === 1 ? "" : "s"}
											saved to the attendee form.
										{:else}
											{" "}The event exists, but the custom form
											still needs to be saved.
										{/if}
									</div>

											<div
												class="border border-[#141414]/12 bg-card p-4
													shadow-sm"
											>
												<h3 class="font-sans text-lg font-medium text-[#141414]">
													Ready to go live?
												</h3>
												<p
													class="mt-2 font-sans text-sm leading-6
														text-[#141414]/66"
												>
													Your draft is saved. Publish it now to make the
													public event page visible and start accepting
													registrations.
												</p>

												<div class="mt-4 flex flex-wrap items-center gap-4">
													<button
														type="button"
														onclick={publishEvent}
														disabled={isPublishing}
														class="rounded-none border border-[#3D6B8C]
															bg-card px-4 py-2 font-sans text-sm font-medium
															text-[#3D6B8C] transition-colors
															hover:bg-[#3D6B8C]/6 disabled:opacity-50"
													>
														{isPublishing
															? "Publishing..."
															: "Publish Event Live"}
													</button>

													{#if publishError}
														<p class="font-sans text-sm text-[#a63c3c]">
															{publishError}
														</p>
													{/if}
												</div>
											</div>
								{/if}
							</div>

							<div class="mt-6 flex flex-wrap gap-3">
								<Button
									type="submit"
									size="lg"
									class="h-14 rounded-none border border-[#3D6B8C]
										bg-card px-6 font-sans text-lg
										text-[#3D6B8C] hover:bg-[#3D6B8C]/6"
									disabled={isSubmitting}
								>
									{isSubmitting
										? "Creating draft..."
										: "Create Draft Event"}
								</Button>

								{#if createdEvent && !createdEvent.formSaved}
									<Button
										type="button"
										size="lg"
										variant="ghost"
										class="h-14 rounded-none border
											border-[#141414]/12 bg-card px-6
											font-sans text-lg text-[#141414]
											hover:bg-[#141414]/4"
										disabled={isRetryingFields}
										onclick={retryFormSave}
									>
										{isRetryingFields
											? "Saving questions..."
											: "Retry Saving Questions"}
									</Button>
								{/if}

								<Button
									type="button"
									size="lg"
									variant="ghost"
									class="h-14 rounded-none border border-[#141414]/12
										bg-card px-6 font-sans text-lg text-[#141414]
										hover:bg-[#141414]/4"
									onclick={resetForm}
								>
									Start Another Draft
								</Button>
							</div>

							<p
								class="mt-4 max-w-3xl font-sans text-base leading-7
									text-[#141414]/66"
							>
								This screen saves the event details through the
								existing event API and then saves the attendee
								form through the event form-fields endpoint.
							</p>
						</div>
					</form>

					<OrganizerEventPreview
						title={eventDetails.title}
						description={eventDetails.description}
						venue={eventDetails.venue}
						startDate={eventDetails.startDate}
						capacity={toFiniteNumber(eventDetails.capacity) || 0}
						visibility={visibilityMode}
						ticketMode={ticketMode}
						price={ticketMode === "paid"
							? toFiniteNumber(ticketPrice) || 0
							: 0}
						fields={previewFields()}
					/>
				</div>
	</div>
</section>
