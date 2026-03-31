<script lang="ts">
	import { page } from "$app/stores";
	import { Button } from "$lib/components/ui/button/index.js";
	import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";

	function isHomeActive(pathname: string): boolean {
		return pathname === "/";
	}

	function isEventsActive(pathname: string): boolean {
		return pathname.startsWith("/events");
	}

	function isOrganizerActive(pathname: string): boolean {
		return pathname.startsWith("/organizer");
	}
</script>

<nav
	class="sticky top-0 z-50 w-full border-b border-[#EAEAEA] bg-white"
>
	<div
		class="mx-auto flex h-16 w-full max-w-[980px] items-center justify-between px-5 sm:px-6"
	>
		<!-- Logo + Nav Links -->
		<div class="flex items-center gap-8">
			<a href="/" class="flex items-center">
				<img
					src="/open_pass_logo.svg"
					alt="Open Pass"
					class="h-7"
				/>
			</a>
			<div class="hidden items-center sm:flex">
				<a
					href="/"
					class={"px-3 py-2 text-sm uppercase tracking-wide transition-colors hover:text-foreground " +
						(isHomeActive($page.url.pathname)
							? "font-semibold text-foreground underline underline-offset-4"
							: "font-medium text-foreground/70")}
				>
					Home
				</a>
				<a
					href="/events"
					class={"px-3 py-2 text-sm uppercase tracking-wide transition-colors hover:text-foreground " +
						(isEventsActive($page.url.pathname)
							? "font-semibold text-foreground underline underline-offset-4"
							: "font-medium text-foreground/70")}
				>
					Events
				</a>
				<a
					href={$page.data.user ? '/organizer/dashboard' : '/organizer/login'}
					class={"px-3 py-2 text-sm uppercase tracking-wide transition-colors hover:text-foreground " +
						(isOrganizerActive($page.url.pathname)
							? "font-semibold text-foreground underline underline-offset-4"
							: "font-medium text-foreground/70")}
				>
					Organize
				</a>
			</div>
		</div>

		<!-- Right Icons -->
		<div class="flex items-center gap-3">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button
							variant="ghost"
							size="icon"
							class="h-10 w-10"
							aria-label="Sign In"
							{...props}
						>
							<svg
								xmlns="http://www.w3.org/2000/svg"
								width="20"
								height="20"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
								stroke-linecap="round"
								stroke-linejoin="round"
							>
								<circle cx="12" cy="8" r="5" />
								<path d="M20 21a8 8 0 0 0-16 0" />
							</svg>
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end" class="w-48 rounded-none border border-[#141414]/14 bg-card font-sans">
					<DropdownMenu.Item>
						{#snippet child({ props })}
							<a href="/login" class="w-full cursor-pointer" {...props}>
								Sign In
							</a>
						{/snippet}
					</DropdownMenu.Item>
					<DropdownMenu.Separator class="bg-[#141414]/14" />
					<DropdownMenu.Item>
						{#snippet child({ props })}
							<a href="/register" class="w-full cursor-pointer" {...props}>
								Create Account
							</a>
						{/snippet}
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>

			<Button
				variant="ghost"
				size="icon"
				class="h-10 w-10"
				aria-label="Toggle theme"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<circle cx="12" cy="12" r="4" />
					<path d="M12 2v2" />
					<path d="M12 20v2" />
					<path d="m4.93 4.93 1.41 1.41" />
					<path d="m17.66 17.66 1.41 1.41" />
					<path d="M2 12h2" />
					<path d="M20 12h2" />
					<path d="m6.34 17.66-1.41 1.41" />
					<path d="m19.07 4.93-1.41 1.41" />
				</svg>
			</Button>
		</div>
	</div>
</nav>
