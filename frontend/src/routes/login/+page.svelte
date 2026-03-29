<script lang="ts">
	import Navbar from "$lib/components/Navbar.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
	import { loginUser } from "$lib/api/auth";

	let email = $state("");
	let password = $state("");
	let submitted = $state(false);
	let errorMsg = $state("");

	const guidelines = [
		"Sign in to track your upcoming events.",
		"Manage your volunteer application status.",
		"Organizers: accessing your dashboard starts here."
	];

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		errorMsg = "";
		if (!email || !password) return;

		try {
			await loginUser(fetch, { email, password });
			submitted = true;
			// Redirect to their dashboard or home
			window.location.href = "/organizer/dashboard"; // For now, just forward organizers or let generic dashboard handle it, but wait, if it's a participant, it might be /dashboard. I will just forward to home for now if logged in.
			window.location.href = "/";
		} catch (err: any) {
			errorMsg = err.message || "Invalid credentials";
		}
	}
</script>

<svelte:head>
	<title>Sign In | Open Pass</title>
	<meta
		name="description"
		content="Sign in to Open Pass to manage your events, tickets, and volunteering."
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
					Sign In
				</p>
				<h1
					class="mt-4 font-serif text-5xl
						leading-tight tracking-tight text-[#141414]
						sm:text-6xl lg:text-7xl"
				>
					<span class="font-normal italic">Welcome</span>
					<span class="font-normal text-[#141414]/65">
						back
					</span>
				</h1>
				<p
					class="mt-6 max-w-xl font-sans text-lg
						leading-8 text-[#141414]/80 sm:text-xl"
				>
					One account, all access. View your passes, manage your organized events, and coordinate with teams.
				</p>

				<div class="mt-8 space-y-3">
					{#each guidelines as guideline}
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
						Don't have an account?
					</p>
					<p
						class="mt-3 font-sans text-lg leading-8
							text-[#141414]/80 sm:text-xl"
					>
						Join Open Pass to get started. Participants and volunteers sign up here.
					</p>
					<div class="mt-4 flex flex-wrap gap-3">
						<Button
							href="/register"
							class="h-12 rounded-none border border-[#3D6B8C]
								bg-card px-5 font-sans text-lg
								text-[#3D6B8C] hover:bg-[#3D6B8C]/6"
						>
							Create Account
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
						Secure Login
					</p>
					<h2
						class="mt-4 font-serif text-5xl leading-tight
							sm:text-6xl lg:text-7xl"
					>
						Sign in to continue
					</h2>
				</div>

				{#if errorMsg}
					<div class="mt-4 border border-red-500/20 bg-red-500/10 p-4 font-sans text-red-600">
						{errorMsg}
					</div>
				{/if}

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
							Email Address
						</label>
						<Input
							id="email"
							name="email"
							type="email"
							autocomplete="email"
							bind:value={email}
							placeholder="you@email.com"
							class="h-14 rounded-none border-[#141414]/14
								bg-card px-4 font-sans text-lg
								text-[#141414] placeholder:text-[#141414]/45"
						/>
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
							Manage your page, registrations, and tickets securely after you sign in.
						</div>

						<Button
							type="submit"
							size="lg"
							class="h-14 w-full rounded-none border
								border-[#3D6B8C] bg-card font-sans
								text-lg text-[#3D6B8C]
								hover:bg-[#3D6B8C]/6"
						>
							Sign In
						</Button>

						{#if submitted}
							<p class="font-sans text-base text-green-600/70">
								Successfully signed in. Redirecting...
							</p>
						{/if}
					</div>
				</form>
			</div>
		</section>
	</main>
</div>
