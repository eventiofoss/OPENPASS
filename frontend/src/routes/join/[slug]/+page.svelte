<script lang="ts">
	import { page } from "$app/stores";
	import Navbar from "$lib/components/Navbar.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { registerUser } from "$lib/api/auth";

	const slug = $derived($page.params.slug);

	let selectedRole: "volunteer" | "participant" = $state("volunteer");
	let fullName = $state("");
	let email = $state("");
	let password = $state("");
	let submitted = $state(false);
	let errorMsg = $state("");

	$effect(() => {
		const param = $page.url.searchParams.get("role");
		if (param === "participant") {
			selectedRole = "participant";
		}
	});

	function canSubmit(): boolean {
		return (
			fullName.trim() !== "" &&
			email.trim() !== "" &&
			password.trim().length >= 8
		);
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		errorMsg = "";
		if (canSubmit()) {
			try {
				await registerUser(fetch, {
					name: fullName.trim(),
					email: email.trim(),
					password: password,
					role: selectedRole
				});
				submitted = true;
			} catch (err: any) {
				errorMsg = err.message || "Something went wrong";
			}
		}
	}
</script>

<svelte:head>
	<title>Join as {selectedRole === "volunteer"
		? "Volunteer"
		: "Participant"} | Open Pass</title>
	<meta
		name="description"
		content="Create an account to join as a {selectedRole}
			for this event on Open Pass."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<main
		class="flex flex-1 items-start px-8 py-8
			sm:px-10 lg:px-12"
	>
		<section
			class="mx-auto grid w-full max-w-6xl
				items-stretch gap-6 lg:grid-cols-2"
		>
			<!-- LEFT: Info Panel -->
			<div
				class="flex h-full flex-col
					border border-[#141414]/14
					bg-card p-7 text-[#141414] sm:p-8
					lg:min-h-[42rem]"
			>
				<p
					class="font-sans text-base
						font-semibold uppercase
						tracking-[0.3em]
						text-[#3D6B8C]"
				>
					Join Event
				</p>
				<h1
					class="mt-4 font-serif text-5xl
						leading-tight tracking-tight
						text-[#141414] sm:text-6xl
						lg:text-7xl"
				>
					<span class="font-normal italic">
						Join
					</span>
					<span
						class="font-normal
							text-[#141414]/65"
					>
						as a {selectedRole}
					</span>
				</h1>
				<p
					class="mt-6 max-w-xl font-sans text-lg
						leading-8 text-[#141414]/80
						sm:text-xl"
				>
					{#if selectedRole === "volunteer"}
						Create an account to help manage
						this event. You'll be able to scan
						QR codes, manage check-ins, and
						coordinate with the organizer.
					{:else}
						Create an account to register for
						events, track your passes, and get
						notified about updates.
					{/if}
				</p>

				<div class="mt-6 space-y-3">
					<div
						class="border-t
							border-[#141414]/14 pt-4"
					>
						<p
							class="font-sans text-base
								font-semibold uppercase
								tracking-[0.2em]
								text-[#3D6B8C]"
						>
							What you'll get
						</p>
						<p
							class="mt-2 font-sans
								text-lg leading-8
								text-[#141414]/80
								sm:text-xl"
						>
							{#if selectedRole === "volunteer"}
								Access to the event
								scanner, check-in tools,
								and team coordination
								dashboard.
							{:else}
								Personalized event feed,
								saved passes, and
								registration history.
							{/if}
						</p>
					</div>

					<div
						class="border-t
							border-[#141414]/14 pt-4"
					>
						<p
							class="font-sans text-base
								font-semibold uppercase
								tracking-[0.2em]
								text-[#3D6B8C]"
						>
							Privacy
						</p>
						<p
							class="mt-2 font-sans
								text-lg leading-8
								text-[#141414]/80
								sm:text-xl"
						>
							Your data stays with the event
							organizer. No third-party
							tracking.
						</p>
					</div>
				</div>

				<div
					class="mt-auto border
						border-[#141414]/14
						bg-[#141414]/3 p-5 pt-6"
				>
					<p
						class="font-sans text-base
							font-semibold uppercase
							tracking-[0.2em]
							text-[#3D6B8C]"
					>
						Just attending?
					</p>
					<p
						class="mt-3 font-sans text-lg
							leading-8
							text-[#141414]/80
							sm:text-xl"
					>
						No account needed for attendees.
						Register directly from the event
						page.
					</p>
					<div
						class="mt-4 flex flex-wrap
							gap-3"
					>
						<Button
							href="/events/{slug}"
							class="h-12 rounded-none
								border
								border-[#3D6B8C]
								bg-card px-5 font-sans
								text-lg
								text-[#3D6B8C]
								hover:bg-[#3D6B8C]/6"
						>
							Back to Event
						</Button>
						<Button
							href="/"
							variant="ghost"
							class="h-12 rounded-none
								border
								border-[#141414]/14
								bg-card px-5 font-sans
								text-lg
								text-[#141414]
								hover:bg-[#141414]/5
								hover:text-[#141414]"
						>
							Back to Discover
						</Button>
					</div>
				</div>
			</div>

			<!-- RIGHT: Form Panel -->
			<div
				class="flex h-full flex-col
					border border-[#141414]/14
					bg-card p-6 text-[#141414] sm:p-7
					lg:min-h-[42rem]"
			>
				{#if submitted}
					<p
						class="font-sans text-base
							font-semibold uppercase
							tracking-[0.3em]
							text-[#3D6B8C]"
					>
						Account Created
					</p>
					<h2
						class="mt-4 font-serif text-5xl
							leading-tight sm:text-6xl
							lg:text-7xl"
					>
						Welcome aboard
					</h2>
					<p
						class="mt-6 max-w-xl font-sans
							text-lg leading-8
							text-[#141414]/80
							sm:text-xl"
					>
						Thanks, {fullName}. Your
						{selectedRole} account has been
						created. Sign in with {email} to
						get started.
					</p>

					<div
						class="mt-8 grid gap-4
							sm:grid-cols-2"
					>
						<div
							class="border
								border-[#141414]/14
								p-4"
						>
							<p
								class="font-sans text-sm
									font-semibold
									uppercase
									tracking-[0.2em]
									text-[#3D6B8C]"
							>
								Account
							</p>
							<p
								class="mt-3 font-sans
									text-lg
									text-[#141414]/80"
							>
								{fullName}
							</p>
							<p
								class="mt-1 font-sans
									text-base
									text-[#141414]/65"
							>
								{email}
							</p>
						</div>
						<div
							class="border
								border-[#141414]/14
								p-4"
						>
							<p
								class="font-sans text-sm
									font-semibold
									uppercase
									tracking-[0.2em]
									text-[#3D6B8C]"
							>
								Role
							</p>
							<p
								class="mt-3 font-sans
									text-lg
									text-[#141414]/80"
							>
								{selectedRole === "volunteer"
									? "Volunteer"
									: "Participant"}
							</p>
						</div>
					</div>

					<div
						class="mt-auto border
							border-[#141414]/14
							bg-[#141414]/3 p-4"
					>
						<p
							class="font-sans text-lg
								leading-7
								text-[#141414]/80
								sm:text-xl"
						>
							Sign in to access your dashboard
							and start helping with this
							event.
						</p>
						<div
							class="mt-5 flex flex-wrap
								gap-3"
						>
							<Button
								href="/organizer/login"
								class="h-12 rounded-none
									border
									border-[#3D6B8C]
									bg-card px-5
									font-sans text-lg
									text-[#3D6B8C]
									hover:bg-[#3D6B8C]/6"
							>
								Sign In
							</Button>
							<Button
								href="/events/{slug}"
								variant="ghost"
								class="h-12 rounded-none
									border
									border-[#141414]/14
									bg-card px-5
									font-sans text-lg
									text-[#141414]
									hover:bg-[#141414]/5"
							>
								Back to Event
							</Button>
						</div>
					</div>
				{:else}
					<p
						class="font-sans text-base
							font-semibold uppercase
							tracking-[0.3em]
							text-[#3D6B8C]"
					>
						Create Account
					</p>
					<h2
						class="mt-4 font-serif text-5xl
							leading-tight sm:text-6xl
							lg:text-7xl"
					>
						{selectedRole === "volunteer"
							? "Volunteer signup"
							: "Join the event"}
					</h2>

					<!-- Role Toggle -->
					<div
						class="mt-6 flex gap-0
							border border-[#141414]/14"
					>
						<button
							type="button"
							class="flex-1 px-4 py-3
								font-sans text-base
								font-medium
								transition-colors
								{selectedRole === 'volunteer'
									? 'bg-[#3D6B8C] text-white'
									: 'bg-card text-[#141414]/70 hover:bg-[#141414]/5'}"
							onclick={() =>
								(selectedRole =
									"volunteer")}
						>
							Volunteer
						</button>
						<button
							type="button"
							class="flex-1 px-4 py-3
								font-sans text-base
								font-medium
								transition-colors
								{selectedRole === 'participant'
									? 'bg-[#3D6B8C] text-white'
									: 'bg-card text-[#141414]/70 hover:bg-[#141414]/5'}"
							onclick={() =>
								(selectedRole =
									"participant")}
						>
							Participant
						</button>
					</div>

					{#if errorMsg}
						<div class="mt-4 border border-red-500/20 bg-red-500/10 p-4 font-sans text-red-600">
							{errorMsg}
						</div>
					{/if}

					<form
						class="mt-6 flex flex-1
							flex-col space-y-4"
						onsubmit={handleSubmit}
					>
						<div class="space-y-2">
							<label
								for="join-name"
								class="font-sans text-lg
									font-medium
									text-[#141414]
									sm:text-xl"
							>
								Full name
							</label>
							<Input
								id="join-name"
								name="name"
								type="text"
								autocomplete="name"
								bind:value={fullName}
								placeholder="Your name"
								class="h-14 rounded-none
									border-[#141414]/14
									bg-card px-4
									font-sans text-lg
									text-[#141414]
									placeholder:text-[#141414]/45"
							/>
						</div>

						<div class="space-y-2">
							<label
								for="join-email"
								class="font-sans text-lg
									font-medium
									text-[#141414]
									sm:text-xl"
							>
								Email
							</label>
							<Input
								id="join-email"
								name="email"
								type="email"
								autocomplete="email"
								bind:value={email}
								placeholder="you@email.com"
								class="h-14 rounded-none
									border-[#141414]/14
									bg-card px-4
									font-sans text-lg
									text-[#141414]
									placeholder:text-[#141414]/45"
							/>
						</div>

						<div class="space-y-2">
							<label
								for="join-password"
								class="font-sans text-lg
									font-medium
									text-[#141414]
									sm:text-xl"
							>
								Password
							</label>
							<Input
								id="join-password"
								name="password"
								type="password"
								autocomplete="new-password"
								bind:value={password}
								placeholder="At least 8 characters"
								class="h-14 rounded-none
									border-[#141414]/14
									bg-card px-4
									font-sans text-lg
									text-[#141414]
									placeholder:text-[#141414]/45"
							/>
							<p
								class="font-sans text-base
									text-[#141414]/70"
							>
								Minimum 8 characters.
							</p>
						</div>

						<div
							class="mt-auto space-y-5
								pt-6"
						>
							<div
								class="border
									border-[#141414]/14
									bg-[#141414]/3 p-4
									font-sans text-lg
									leading-8
									text-[#141414]/75
									sm:text-xl"
							>
								{#if selectedRole === "volunteer"}
									After creating your
									account, the organizer
									will grant you access to
									event tools.
								{:else}
									Your account lets you
									track your events and
									passes in one place.
								{/if}
							</div>

							<div
								class="flex flex-wrap
									gap-3"
							>
								<Button
									type="submit"
									class="h-14 flex-1
										rounded-none
										border
										border-[#3D6B8C]
										bg-card px-5
										font-sans
										text-lg
										text-[#3D6B8C]
										hover:bg-[#3D6B8C]/6"
									disabled={!canSubmit()}
								>
									Create{" "}
									{selectedRole ===
									"volunteer"
										? "Volunteer"
										: "Participant"}
									{" "}Account
								</Button>
								<Button
									href="/events/{slug}"
									variant="ghost"
									class="h-14
										rounded-none
										border
										border-[#141414]/14
										bg-card px-5
										font-sans
										text-lg
										text-[#141414]
										hover:bg-[#141414]/5"
								>
									Cancel
								</Button>
							</div>
						</div>
					</form>
				{/if}
			</div>
		</section>
	</main>
</div>
