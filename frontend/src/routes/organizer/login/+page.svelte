<script lang="ts">
	import Navbar from "$lib/components/Navbar.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { goto, invalidateAll } from "$app/navigation";
	import { loginUser } from "$lib/api/auth";

	let email = $state("");
	let password = $state("");
	let submitted = $state(false);
	let errorMsg = $state("");

	const organizerGuidelines = [
		"Only organizers can sign in here.",
		"New accounts need approval first.",
	];

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		errorMsg = "";
		if (!email || !password) {
			errorMsg = "Email and password are required.";
			submitted = true;
			return;
		}

		try {
			await loginUser(fetch, { email, password });
			submitted = true;
			await invalidateAll();
			await goto("/organizer/dashboard");
		} catch (err: unknown) {
			errorMsg = err instanceof Error
				? err.message
				: "Invalid credentials";
			submitted = true;
		}
	}
</script>

<svelte:head>
	<title>Organizer Login | Open Pass</title>
	<meta
		name="description"
		content="Organizer-only login for Open Pass event management,
		check-ins, and attendee operations."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<main
		class="flex flex-1 items-center px-8 py-8
			sm:px-10 lg:px-12"
	>
		<section
			class="mx-auto grid w-full max-w-6xl items-stretch gap-6
				lg:grid-cols-2"
		>
			<div
				class="flex h-full flex-col border border-[#141414]/14
					bg-card p-7 text-[#141414] sm:p-8
					lg:min-h-[44rem]"
			>
				<p
					class="font-sans text-base font-semibold uppercase
						tracking-[0.3em] text-[#3D6B8C]"
				>
					Organizer Login
				</p>
				<h1
					class="mt-4 font-serif text-5xl
						leading-tight tracking-tight text-[#141414]
						sm:text-6xl lg:text-7xl"
				>
					<span class="font-normal italic">Sign in</span>
					<span class="font-normal text-[#141414]/65">
						to run your events
					</span>
				</h1>
				<p
					class="mt-6 max-w-xl font-sans text-lg
						leading-8 text-[#141414]/80 sm:text-xl"
				>
					Open your dashboard, manage registrations, and check
					people in.
				</p>

				<div class="mt-8 space-y-3">
					{#each organizerGuidelines as guideline}
						<div
							class="border-t border-[#141414]/14 pt-4
								font-sans text-lg text-[#141414]/80
								sm:text-xl"
						>
							{guideline}
						</div>
					{/each}
				</div>

				<div
					class="mt-auto border border-[#141414]/14
						bg-[#141414]/3 p-5 pt-6"
				>
					<p
						class="font-sans text-base font-semibold uppercase
							tracking-[0.2em] text-[#3D6B8C]"
					>
						Need approval first?
					</p>
					<p
						class="mt-3 font-sans text-lg leading-8
							text-[#141414]/80 sm:text-xl"
					>
						Applied already? Come back after approval.
					</p>
					<div class="mt-4 flex flex-wrap gap-3">
						<Button
							href="/organizer/register"
							class="h-12 rounded-none border border-[#3D6B8C]
								bg-card px-5 font-sans text-lg
								text-[#3D6B8C] hover:bg-[#3D6B8C]/6"
						>
							Create Organizer Account
						</Button>
						<Button
							href="/"
							variant="ghost"
							class="h-12 rounded-none border border-[#141414]/14
								bg-card px-5 font-sans text-lg text-[#141414]
								hover:bg-[#141414]/5 hover:text-[#141414]"
						>
							Back to Discover
						</Button>
					</div>
				</div>
			</div>

			<div
				class="flex h-full flex-col border border-[#141414]/14
					bg-card p-7 text-[#141414] sm:p-8
					lg:min-h-[44rem]"
			>
				<div>
					<p
						class="font-sans text-base font-semibold uppercase
							tracking-[0.3em] text-[#3D6B8C]"
					>
						Sign In
					</p>
					<h2
						class="mt-4 font-serif text-5xl leading-tight
							sm:text-6xl lg:text-7xl"
					>
						Organizer login
					</h2>
				</div>

				<form
					class="mt-8 flex flex-1 flex-col space-y-5"
					onsubmit={handleSubmit}
				>
					<div class="space-y-2">
						<label
							for="email"
							class="font-sans text-lg font-medium text-[#141414]
								sm:text-xl"
						>
							Work email
						</label>
						<Input
							id="email"
							name="email"
							type="email"
							autocomplete="email"
							bind:value={email}
							placeholder="organizer@studio.com"
							class="h-14 rounded-none border-[#141414]/14
								bg-card px-4 font-sans text-lg
								text-[#141414] placeholder:text-[#141414]/45"
						/>
						<p class="font-sans text-base text-[#141414]/70">
							Use the email for your organizer account.
						</p>
					</div>

					<div class="space-y-2">
						<div class="flex items-center justify-between gap-4">
							<label
								for="password"
								class="font-sans text-lg font-medium text-[#141414]
									sm:text-xl"
							>
								Password
							</label>
							<a
								href="/organizer/register"
								class="font-sans text-base text-[#3D6B8C]
									underline-offset-4 transition-colors
									hover:underline"
							>
								Need access?
							</a>
						</div>
						<Input
							id="password"
							name="password"
							type="password"
							autocomplete="current-password"
							bind:value={password}
							placeholder="Enter your password"
							class="h-14 rounded-none border-[#141414]/14
								bg-card px-4 font-sans text-lg
								text-[#141414] placeholder:text-[#141414]/45"
						/>
					</div>

					<div class="mt-auto space-y-5 pt-6">
						<div
							class="border border-[#141414]/14 bg-[#141414]/3 p-4
								font-sans text-lg leading-8 text-[#141414]/75
								sm:text-xl"
						>
							Manage your page, registrations, and check-ins after
							you sign in.
						</div>

						<Button
							type="submit"
							size="lg"
							class="h-14 w-full rounded-none border
								border-[#3D6B8C] bg-card font-sans
								text-lg text-[#3D6B8C]
								hover:bg-[#3D6B8C]/6"
						>
							Continue as Organizer
						</Button>

						{#if submitted}
							<p class="font-sans text-base text-[#141414]/70">
								{errorMsg || "Signed in successfully. Redirecting..."}
							</p>
						{/if}
					</div>
				</form>
			</div>
		</section>
	</main>
</div>
