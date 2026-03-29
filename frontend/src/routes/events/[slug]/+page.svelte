<script lang="ts">
	import { page } from "$app/stores";
	import Navbar from "$lib/components/Navbar.svelte";
	import Footer from "$lib/components/Footer.svelte";
	import { Badge } from "$lib/components/ui/badge/index.js";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const event = $derived({
		slug: data.event.slug,
		title: data.event.title,
		description: data.event.description,
		poster_url: data.event.poster_url,
		venue: data.event.venue,
		startDate: data.event.start_date,
		capacity: data.event.capacity,
		totalRegistered: data.event.total_registered,
		price: data.event.price,
		status: data.event.status,
		organizerName: data.event.organizer_name,
		organizerEmail: data.event.organizer_email,
	});

	// For the QR Code absolute URL
	const absUrl = $derived($page.url.origin + '/events/' + event.slug);
	const qrCodeUrl = $derived(`https://api.qrserver.com/v1/create-qr-code/?size=150x150&data=${encodeURIComponent(absUrl)}`);

	const volunteers = [
		{ name: "Asha M.", role: "Registration Desk" },
		{ name: "Rajeev K.", role: "Stage Coordinator" },
		{ name: "Priya S.", role: "Photography" },
	];

	let regName = $state("");
	let regEmail = $state("");
	let regSubmitted = $state(false);

	function formatDate(iso: string): string {
		const d = new Date(iso);
		return d.toLocaleDateString("en-IN", {
			weekday: "long",
			day: "numeric",
			month: "long",
			year: "numeric",
		});
	}

	function formatTime(iso: string): string {
		const d = new Date(iso);
		return d.toLocaleTimeString("en-IN", {
			hour: "2-digit",
			minute: "2-digit",
			hour12: true,
		});
	}

	function capacityPct(): number {
		if (event.capacity <= 0) return 0;
		return Math.min(
			100,
			Math.round(
				(event.totalRegistered / event.capacity) * 100
			)
		);
	}

	function handleRegister(e: SubmitEvent) {
		e.preventDefault();
		regSubmitted = true;
	}
</script>

<svelte:head>
	<title>{event.title} | Open Pass</title>
	<meta
		name="description"
		content={event.description.slice(0, 155)}
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<main class="flex-1">
		<!-- Poster / Hero Banner -->
		<section class="px-8 pt-8 sm:px-10 lg:px-12">
			<div class="mx-auto max-w-7xl">
				<div
					class="relative aspect-[3/1] w-full
						overflow-hidden
						border border-[#141414]/14
						bg-[#141414]/5"
				>
					{#if event.poster_url}
						<!-- Added Poster Image -->
						<img
							src={event.poster_url}
							alt="{event.title} Poster"
							class="absolute inset-0 h-full w-full object-cover"
						/>
						<!-- Gradient overlay to ensure text legibility -->
						<div class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/30 to-transparent"></div>
					{:else}
						<!-- Fallback Gradient -->
						<div
							class="absolute inset-0"
							style="background: linear-gradient(
								135deg, #3D6B8C 0%, #2a4d66 40%,
								#1a3344 100%
							);"
						></div>
						<!-- Decorative circles -->
						<div
							class="absolute -right-20 -top-20
								h-72 w-72 rounded-full
								bg-white/5"
						></div>
						<div
							class="absolute -bottom-16 left-1/4
								h-56 w-56 rounded-full
								bg-white/5"
						></div>
					{/if}

					<div
						class="absolute inset-0 flex
							flex-col items-start
							justify-end p-8 sm:p-12"
					>
						<Badge
							variant="secondary"
							class="mb-4 !rounded-none
								px-3 py-1 text-sm
								font-semibold uppercase
								tracking-widest"
						>
							{event.status === "full"
								? "Sold Out"
								: "Live"}
						</Badge>
						<h1
							class="font-serif text-4xl
								leading-tight
								text-white sm:text-5xl
								lg:text-6xl"
						>
							{event.title}
						</h1>
					</div>
				</div>
			</div>
		</section>

		<!-- Two-Column Content -->
		<section class="px-8 py-8 sm:px-10 lg:px-12">
			<div class="mx-auto max-w-7xl">
				<div
					class="grid grid-cols-1 items-start
						gap-6 lg:grid-cols-5"
				>
					<!-- LEFT: Event Info (3 cols) -->
					<div
						class="flex flex-col gap-6
							lg:col-span-3"
					>
						<!-- About -->
						<div
							class="border border-[#141414]/14
								bg-card p-7 sm:p-8"
						>
							<p
								class="font-sans text-base
									font-semibold uppercase
									tracking-[0.3em]
									text-[#3D6B8C]"
							>
								About this Event
							</p>
							<div
								class="mt-6 space-y-4
									font-sans text-lg
									leading-8
									text-[#141414]/80
									sm:text-xl"
							>
								{#each event.description.split("\n\n") as para}
									<p>{para}</p>
								{/each}
							</div>
						</div>

						<!-- Details Grid -->
						<div
							class="grid grid-cols-1 gap-6
								sm:grid-cols-2"
						>
							<!-- Date & Time -->
							<div
								class="border
									border-[#141414]/14
									bg-card p-6"
							>
								<p
									class="font-sans text-sm
										font-semibold
										uppercase
										tracking-[0.2em]
										text-[#3D6B8C]"
								>
									Date & Time
								</p>
								<p
									class="mt-3 font-sans
										text-lg
										text-[#141414]"
								>
									{formatDate(
										event.startDate
									)}
								</p>
								<p
									class="mt-1 font-sans
										text-base
										text-[#141414]/65"
								>
									{formatTime(
										event.startDate
									)}
								</p>
							</div>

							<!-- Location -->
							<div
								class="border
									border-[#141414]/14
									bg-card p-6"
							>
								<p
									class="font-sans text-sm
										font-semibold
										uppercase
										tracking-[0.2em]
										text-[#3D6B8C]"
								>
									Location
								</p>
								<p
									class="mt-3 font-sans
										text-lg
										text-[#141414]"
								>
									{event.venue}
								</p>
								<button
									class="mt-1 font-sans
										text-base
										text-[#3D6B8C]
										underline-offset-4
										hover:underline"
								>
									View on Map
								</button>
							</div>

							<!-- Price -->
							<div
								class="border
									border-[#141414]/14
									bg-card p-6"
							>
								<p
									class="font-sans text-sm
										font-semibold
										uppercase
										tracking-[0.2em]
										text-[#3D6B8C]"
								>
									Price
								</p>
								<p
									class="mt-3 font-serif
										text-3xl
										text-[#141414]"
								>
									{event.price === 0
										? "Free"
										: `₹${event.price}`}
								</p>
							</div>

							<!-- Capacity -->
							<div
								class="border
									border-[#141414]/14
									bg-card p-6"
							>
								<p
									class="font-sans text-sm
										font-semibold
										uppercase
										tracking-[0.2em]
										text-[#3D6B8C]"
								>
									Capacity
								</p>
								<div class="mt-3 space-y-2">
									<div
										class="flex
											items-center
											justify-between"
									>
										<span
											class="font-sans
												text-lg
												text-[#141414]"
										>
											{event.totalRegistered}
											/ {event.capacity}
										</span>
										<span
											class="font-sans
												text-base
												text-[#141414]/65"
										>
											{capacityPct()}%
										</span>
									</div>
									<div
										class="h-2 w-full
											bg-[#141414]/8"
									>
										<div
											class="h-full
												transition-all
												duration-500"
											style="width: {capacityPct()}%;
												background-color: {capacityPct() >= 90
													? '#e11d48'
													: '#3D6B8C'};"
										></div>
									</div>
								</div>
							</div>
						</div>
					</div>

					<!-- RIGHT: Action Panel (2 cols) -->
					<div
						class="flex flex-col gap-6
							lg:col-span-2"
					>
						<!-- Volunteers -->
						<div
							class="border border-[#141414]/14
								bg-card p-7 sm:p-8"
						>
							<p
								class="font-sans text-base
									font-semibold uppercase
									tracking-[0.3em]
									text-[#3D6B8C]"
							>
								Volunteers
							</p>
							<p
								class="mt-2 font-sans
									text-base
									text-[#141414]/65"
							>
								{volunteers.length} people
								helping out
							</p>

							<div class="mt-5 space-y-3">
								{#each volunteers as vol}
									<div
										class="flex
											items-center
											gap-3
											border-t
											border-[#141414]/8
											pt-3"
									>
										<div
											class="flex h-9
												w-9
												items-center
												justify-center
												bg-[#3D6B8C]/10
												font-sans
												text-sm
												font-semibold
												text-[#3D6B8C]"
										>
											{vol.name
												.charAt(0)
												.toUpperCase()}
										</div>
										<div>
											<p
												class="font-sans
													text-base
													font-medium
													text-[#141414]"
											>
												{vol.name}
											</p>
											<p
												class="font-sans
													text-sm
													text-[#141414]/60"
											>
												{vol.role}
											</p>
										</div>
									</div>
								{/each}
							</div>

							<Button
								href="/join/{event.slug}?role=volunteer"
								class="mt-6 h-12 w-full
									rounded-none border
									border-[#3D6B8C]
									bg-card px-5 font-sans
									text-lg text-[#3D6B8C]
									hover:bg-[#3D6B8C]/6"
							>
								Join as Volunteer
							</Button>
						</div>

						<!-- Register Section -->
						<div
							class="border border-[#141414]/14
								bg-card p-7 sm:p-8"
						>
							<div class="mb-6 flex flex-col items-center justify-between sm:flex-row">
								<p
									class="font-sans text-base
										font-semibold uppercase
										tracking-[0.3em]
										text-[#3D6B8C]"
								>
									Registration
								</p>
								
								<!-- Share Event QR Code -->
								<div class="mt-4 flex flex-col items-center sm:mt-0">
									<img src={qrCodeUrl} alt="Event QR Code" class="h-16 w-16 border border-[#141414]/10 bg-white p-1" />
									<span class="mt-1 text-xs text-[#141414]/50">Scan to View</span>
								</div>
							</div>

							{#if regSubmitted}
								<div class="mt-6">
									<p
										class="font-serif
											text-3xl
											text-[#141414]"
									>
										You're in!
									</p>
									<p
										class="mt-3
											font-sans
											text-lg
											leading-8
											text-[#141414]/80"
									>
										Check {regEmail} for
										your pass and QR code.
									</p>
								</div>
							{:else if event.status === "full"}
								<div class="mt-6">
									<p
										class="font-sans
											text-lg
											text-[#141414]/80"
									>
										This event is full.
										Check back later or
										contact the organizer.
									</p>
								</div>
							{:else}
								<form
									class="mt-6 space-y-4"
									onsubmit={handleRegister}
								>
									<div class="space-y-2">
										<label
											for="reg-name"
											class="font-sans
												text-lg
												font-medium
												text-[#141414]"
										>
											Full name
										</label>
										<Input
											id="reg-name"
											name="name"
											type="text"
											autocomplete="name"
											bind:value={regName}
											placeholder="Your name"
											class="h-14
												rounded-none
												border-[#141414]/14
												bg-card px-4
												font-sans
												text-lg
												text-[#141414]
												placeholder:text-[#141414]/45"
										/>
									</div>

									<div class="space-y-2">
										<label
											for="reg-email"
											class="font-sans
												text-lg
												font-medium
												text-[#141414]"
										>
											Email
										</label>
										<Input
											id="reg-email"
											name="email"
											type="email"
											autocomplete="email"
											bind:value={regEmail}
											placeholder="you@email.com"
											class="h-14
												rounded-none
												border-[#141414]/14
												bg-card px-4
												font-sans
												text-lg
												text-[#141414]
												placeholder:text-[#141414]/45"
										/>
									</div>

									<Button
										type="submit"
										class="h-14 w-full
											rounded-none
											bg-foreground
											font-sans
											text-lg
											text-background
											hover:bg-foreground/90"
										disabled={regName.trim() === "" ||
											regEmail.trim() === ""}
									>
										Register for Free
									</Button>

									<p
										class="text-center
											font-sans
											text-sm
											text-[#141414]/60"
									>
										No account needed —
										you'll receive a QR
										pass via email.
									</p>
								</form>
							{/if}
						</div>

						<!-- Contact -->
						<div
							class="border border-[#141414]/14
								bg-[#141414]/3 p-7
								sm:p-8"
						>
							<p
								class="font-sans text-base
									font-semibold uppercase
									tracking-[0.3em]
									text-[#3D6B8C]"
							>
								Contact
							</p>
							<p
								class="mt-4 font-sans
									text-lg text-[#141414]"
							>
								{event.organizerName}
							</p>
							<a
								href="mailto:{event.organizerEmail}"
								class="mt-1 block font-sans
									text-base
									text-[#3D6B8C]
									underline-offset-4
									transition-colors
									hover:underline"
							>
								{event.organizerEmail}
							</a>
							<div
								class="mt-5 flex flex-wrap
									gap-3"
							>
								<Button
									href="/"
									variant="ghost"
									class="h-12 rounded-none
										border
										border-[#141414]/14
										bg-card px-5
										font-sans text-base
										text-[#141414]
										hover:bg-[#141414]/5
										hover:text-[#141414]"
								>
									Back to Discover
								</Button>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>
	</main>

	<Footer />
</div>
