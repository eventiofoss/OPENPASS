<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
</script>

<nav
	class="sticky top-0 z-50 border-b border-border
		bg-background/95 backdrop-blur-sm"
>
	<div
		class="mx-auto flex h-18 max-w-7xl items-center
			justify-between px-8 sm:px-10 lg:px-12"
	>
		<!-- Logo + Nav Links -->
		<div class="flex items-center gap-10">
			<a href="/" class="flex items-center">
				<img
					src="/open_pass_logo.svg"
					alt="Open Pass"
					class="h-8"
				/>
			</a>
			<div class="hidden items-center sm:flex">
				<a
					href="/"
					class="px-4 py-2 text-sm font-semibold
						uppercase tracking-widest
						text-foreground/70
						transition-colors
						hover:text-foreground
						underline"
				>
					Home
				</a>
				<a
					href="/events"
					class="px-4 py-2 text-sm font-semibold
						uppercase tracking-widest
						text-foreground/70
						transition-colors
						hover:text-foreground"
				>
					Events
				</a>
				<a
					href="/organizer/dashboard"
					class="px-4 py-2 text-sm font-semibold
						uppercase tracking-widest text-accent
						transition-colors
						hover:text-accent/80"
				>
					Organize
				</a>
			</div>
		</div>

		<!-- Right Icons -->
		<div class="flex items-center gap-3">
			<Button
				variant="ghost"
				size="icon"
				class="h-10 w-10"
				aria-label="Profile"
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
