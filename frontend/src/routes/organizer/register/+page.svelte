<script lang="ts">
	import Navbar from "$lib/components/Navbar.svelte";
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";

	const steps = [
		{
			eyebrow: "Step 1",
			title: "Create your account",
			copy: "Start with the account fields the current backend expects.",
		},
		{
			eyebrow: "Step 2",
			title: "Tell us about your organization",
			copy: "Share the basics reviewers need before approving organizer access.",
		},
		{
			eyebrow: "Step 3",
			title: "Finish your review packet",
			copy: "Add your use case and ID upload so the application is ready for review.",
		},
	];

	const checklist = [
		{
			title: "Personal info",
			copy: "Name, email, and password.",
		},
		{
			title: "Organization info",
			copy: "Organization, role, and location.",
		},
		{
			title: "Review details",
			copy: "Use case and one ID upload.",
		},
	];

	let currentStep = $state(0);
	let submitted = $state(false);
	let fullName = $state("");
	let email = $state("");
	let password = $state("");
	let organizationName = $state("");
	let organizerRole = $state("");
	let location = $state("");
	let useCase = $state("");
	let idFiles = $state<FileList | undefined>(undefined);
	let confirmsReview = $state(false);

	function canContinue(stepIndex: number) {
		if (stepIndex === 0) {
			return (
				fullName.trim() !== "" &&
				email.trim() !== "" &&
				password.trim().length >= 8
			);
		}

		if (stepIndex === 1) {
			return (
				organizationName.trim() !== "" &&
				organizerRole.trim() !== "" &&
				location.trim() !== ""
			);
		}

		return (
			useCase.trim() !== "" &&
			(idFiles?.length ?? 0) > 0 &&
			confirmsReview
		);
	}

	function nextStep() {
		if (currentStep < steps.length - 1 && canContinue(currentStep)) {
			currentStep += 1;
		}
	}

	function previousStep() {
		if (currentStep > 0) {
			currentStep -= 1;
		}
	}

	function submitApplication(event: SubmitEvent) {
		event.preventDefault();

		if (canContinue(steps.length - 1)) {
			submitted = true;
		}
	}

	function selectedIdLabel() {
		return idFiles?.[0]?.name ?? "Upload a government-issued ID";
	}
</script>

<svelte:head>
	<title>Create Organizer Account | Open Pass</title>
	<meta
		name="description"
		content="Apply for an organizer account on Open Pass with account details, organization information, and review documents."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<main
		class="flex flex-1 items-start px-8 py-8
			sm:px-10 lg:px-12"
	>
		<section
			class="mx-auto grid w-full max-w-6xl items-stretch gap-6
				lg:grid-cols-2"
		>
			<div
				class="flex h-full flex-col border border-[#141414]/14
					bg-card p-7 text-[#141414] sm:p-8
					lg:min-h-[42rem]"
			>
				<p
					class="font-sans text-base font-semibold uppercase
						tracking-[0.3em] text-[#3D6B8C]"
				>
					Organizer Register
				</p>
				<h1
					class="mt-4 font-serif text-5xl
						leading-tight tracking-tight text-[#141414]
						sm:text-6xl lg:text-7xl"
				>
					<span class="font-normal italic">Apply</span>
					<span class="font-normal text-[#141414]/65">
						to host with Open Pass
					</span>
				</h1>
				<p
					class="mt-6 max-w-xl font-sans text-lg
						leading-8 text-[#141414]/80 sm:text-xl"
				>
					Create your account and send the details needed for
					approval.
				</p>

				<div class="mt-6 space-y-3">
					{#each checklist as item}
						<div class="border-t border-[#141414]/14 pt-4">
							<p
								class="font-sans text-base font-semibold
									uppercase tracking-[0.2em] text-[#3D6B8C]"
							>
								{item.title}
							</p>
							<p
								class="mt-2 font-sans text-lg
									leading-8 text-[#141414]/80
									sm:text-xl"
							>
								{item.copy}
							</p>
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
						Already approved?
					</p>
					<p
						class="mt-3 font-sans text-lg leading-8
							text-[#141414]/80 sm:text-xl"
					>
						Go back to login if you're already approved.
					</p>
					<div class="mt-4 flex flex-wrap gap-3">
						<Button
							href="/organizer/login"
							class="h-12 rounded-none border border-[#3D6B8C]
								bg-card px-5 font-sans text-lg
								text-[#3D6B8C] hover:bg-[#3D6B8C]/6"
						>
							Go to Organizer Login
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
					bg-card p-6 text-[#141414] sm:p-7
					lg:min-h-[42rem]"
			>
				{#if submitted}
					<p
						class="font-sans text-base font-semibold uppercase
							tracking-[0.3em] text-[#3D6B8C]"
					>
						Pending Review
					</p>
					<h2
						class="mt-4 font-serif text-5xl leading-tight
							sm:text-6xl lg:text-7xl"
					>
						Application received
					</h2>
					<p
						class="mt-6 max-w-xl font-sans text-lg
							leading-8 text-[#141414]/80 sm:text-xl"
					>
						Thanks, {fullName}. Your organizer application is now
						in review. Once approved, sign in with {email}.
					</p>

					<div class="mt-8 grid gap-4 sm:grid-cols-2">
						<div class="border border-[#141414]/14 p-4">
							<p
								class="font-sans text-sm font-semibold uppercase
									tracking-[0.2em] text-[#3D6B8C]"
							>
								Account
							</p>
							<p class="mt-3 font-sans text-lg text-[#141414]/80">
								{fullName}
							</p>
							<p class="mt-1 font-sans text-base text-[#141414]/65">
								{email}
							</p>
						</div>
						<div class="border border-[#141414]/14 p-4">
							<p
								class="font-sans text-sm font-semibold uppercase
									tracking-[0.2em] text-[#3D6B8C]"
							>
								Organization
							</p>
							<p class="mt-3 font-sans text-lg text-[#141414]/80">
								{organizationName}
							</p>
							<p class="mt-1 font-sans text-base text-[#141414]/65">
								{organizerRole} in {location}
							</p>
						</div>
					</div>

					<div
						class="mt-auto border border-[#141414]/14
							bg-[#141414]/3 p-4"
					>
						<p
							class="font-sans text-lg leading-7
								text-[#141414]/80 sm:text-xl"
						>
							Check your organizer email, then come back here to
							sign in after approval.
						</p>
						<div class="mt-5 flex flex-wrap gap-3">
							<Button
								href="/organizer/login"
								class="h-12 rounded-none border
									border-[#3D6B8C] bg-card px-5
									font-sans text-lg text-[#3D6B8C]
									hover:bg-[#3D6B8C]/6"
							>
								Return to Login
							</Button>
							<Button
								href="/"
								variant="ghost"
								class="h-12 rounded-none border
									border-[#141414]/14 bg-card px-5 font-sans
									text-lg text-[#141414]
									hover:bg-[#141414]/5"
							>
								Back to Home
							</Button>
						</div>
					</div>
				{:else}
					<p
						class="font-sans text-base font-semibold uppercase
							tracking-[0.3em] text-[#3D6B8C]"
					>
						{steps[currentStep].eyebrow}
					</p>
					<h2
						class="mt-4 font-serif text-5xl leading-tight
							sm:text-6xl lg:text-7xl"
					>
						{steps[currentStep].title}
					</h2>
					<p
						class="mt-6 max-w-xl font-sans text-lg
							leading-8 text-[#141414]/80 sm:text-xl"
					>
						{steps[currentStep].copy}
					</p>

					<div class="mt-8 flex gap-3">
						{#each steps as _, index}
							<div
								class={`h-1 flex-1 ${
									index <= currentStep
										? "bg-[#3D6B8C]"
										: "bg-[#141414]/10"
								}`}
							></div>
						{/each}
					</div>

					<form
						class="mt-8 flex flex-1 flex-col"
						onsubmit={submitApplication}
					>
						<div class="space-y-4">
							{#if currentStep === 0}
								<div class="space-y-2">
									<label
										for="name"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										Full name
									</label>
									<Input
										id="name"
										name="name"
										type="text"
										autocomplete="name"
										bind:value={fullName}
										placeholder="Alicia Thomas"
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-lg
											text-[#141414]
											placeholder:text-[#141414]/45"
									/>
								</div>

								<div class="space-y-2">
									<label
										for="email"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
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
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-lg
											text-[#141414]
											placeholder:text-[#141414]/45"
									/>
								</div>

								<div class="space-y-2">
									<label
										for="password"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										Password
									</label>
									<Input
										id="password"
										name="password"
										type="password"
										autocomplete="new-password"
										bind:value={password}
										placeholder="At least 8 characters"
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-lg
											text-[#141414]
											placeholder:text-[#141414]/45"
									/>
									<p class="font-sans text-base text-[#141414]/70">
										The current backend requires name, email,
										and a password with at least 8 characters.
									</p>
								</div>
							{/if}

							{#if currentStep === 1}
								<div class="space-y-2">
									<label
										for="organization-name"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										Organization name
									</label>
									<Input
										id="organization-name"
										name="organization_name"
										type="text"
										bind:value={organizationName}
										placeholder="Open Arts Collective"
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-lg
											text-[#141414]
											placeholder:text-[#141414]/45"
									/>
								</div>

								<div class="space-y-2">
									<label
										for="organizer-role"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										Your role
									</label>
									<Input
										id="organizer-role"
										name="organizer_role"
										type="text"
										bind:value={organizerRole}
										placeholder="Founder, producer, community lead..."
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-lg
											text-[#141414]
											placeholder:text-[#141414]/45"
									/>
								</div>

								<div class="space-y-2">
									<label
										for="location"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										Primary location
									</label>
									<Input
										id="location"
										name="location"
										type="text"
										bind:value={location}
										placeholder="Kochi, India"
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-lg
											text-[#141414]
											placeholder:text-[#141414]/45"
									/>
								</div>
							{/if}

							{#if currentStep === 2}
								<div class="space-y-2">
									<label
										for="use-case"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										How will you use Open Pass?
									</label>
									<textarea
										id="use-case"
										name="use_case"
										bind:value={useCase}
										placeholder="Tell us what kind of events you run and what you need from the platform."
										class="min-h-40 w-full rounded-none
											border border-[#141414]/14
											bg-card px-4 py-3 font-sans
											text-lg leading-8 text-[#141414]
											outline-none transition-[color,box-shadow]
											placeholder:text-[#141414]/45
											focus-visible:border-[#3D6B8C]
											focus-visible:ring-3
											focus-visible:ring-[#3D6B8C]/20"
									></textarea>
								</div>

								<div class="space-y-2">
									<label
										for="government-id"
										class="font-sans text-lg font-medium
											text-[#141414] sm:text-xl"
									>
										Government-issued ID
									</label>
									<Input
										id="government-id"
										name="government_id"
										type="file"
										accept=".pdf,image/png,image/jpeg"
										bind:files={idFiles}
										class="h-14 rounded-none
											border-[#141414]/14 bg-card
											px-4 font-sans text-base
											text-[#141414]"
									/>
									<p
										class="mt-3 font-sans text-base
											text-[#141414]/70"
									>
										{selectedIdLabel()}
									</p>
								</div>

								<label
									class="flex items-start gap-3 border
										border-[#141414]/14 bg-[#141414]/3
										p-4 font-sans text-base leading-7
										text-[#141414]/75"
								>
									<input
										type="checkbox"
										bind:checked={confirmsReview}
										class="mt-1 h-4 w-4 rounded-none
											accent-[#3D6B8C]"
									/>
									<span>
										I understand this application should stay
										in pending review until organizer access
										is approved.
									</span>
								</label>
							{/if}
						</div>

						<div class="mt-auto flex flex-wrap gap-3 pt-8">
							{#if currentStep > 0}
								<Button
									type="button"
									variant="ghost"
									class="h-12 rounded-none border
										border-[#141414]/14 bg-card px-5 font-sans
										text-lg text-[#141414]
										hover:bg-[#141414]/5"
									onclick={previousStep}
								>
									Back
								</Button>
							{:else}
								<Button
									href="/organizer/login"
									variant="ghost"
									class="h-12 rounded-none border
										border-[#141414]/14 bg-card px-5 font-sans
										text-lg text-[#141414]
										hover:bg-[#141414]/5"
								>
									Already have access?
								</Button>
							{/if}

							{#if currentStep < steps.length - 1}
								<Button
									type="button"
									class="h-12 rounded-none border
										border-[#3D6B8C] bg-card px-5
										font-sans text-lg text-[#3D6B8C]
										hover:bg-[#3D6B8C]/6"
									onclick={nextStep}
									disabled={!canContinue(currentStep)}
								>
									Continue
								</Button>
							{:else}
								<Button
									type="submit"
									class="h-12 rounded-none border
										border-[#3D6B8C] bg-card px-5
										font-sans text-lg text-[#3D6B8C]
										hover:bg-[#3D6B8C]/6"
									disabled={!canContinue(currentStep)}
								>
									Submit for Review
								</Button>
							{/if}
						</div>
					</form>
				{/if}
			</div>
		</section>
	</main>
</div>
