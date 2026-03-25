<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
</script>

<section class="px-4 pb-8 pt-10 sm:px-6">
	<div class="mx-auto max-w-6xl">
		<!-- Heading -->
		<h1
			class="font-serif text-4xl leading-tight
				tracking-tight sm:text-5xl"
		>
			<span class="font-normal italic">Find</span>
			<span class="font-normal">
				Whats Happening
			</span>
		</h1>

		<!-- Location -->
		<div class="mt-2 flex items-center gap-1.5">
			<span class="text-sm text-muted-foreground">
				Near You,
			</span>
			<span
				class="text-sm font-medium text-foreground"
			>
				Karukutty, IN
			</span>
			<button
				class="ml-0.5 text-accent
					transition-colors hover:text-accent/80"
				aria-label="Change location"
			>
				<svg
					xmlns="http://www.w3.org/2000/svg"
					width="14"
					height="14"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				>
					<path
						d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"
					/>
					<path d="m15 5 4 4" />
				</svg>
			</button>
		</div>

		<!-- Search Bar -->
		<div
			class="mt-6 flex items-center gap-2 rounded-lg
				border border-border bg-card px-3 py-2
				shadow-sm"
		>
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="18"
				height="18"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				class="shrink-0 text-muted-foreground"
			>
				<circle cx="11" cy="11" r="8" />
				<path d="m21 21-4.3-4.3" />
			</svg>
			<Input
				type="text"
				placeholder="Search Events, Organizers, or Topics..."
				class="!border-0 !bg-transparent !shadow-none
					!outline-none !ring-0 focus-visible:!ring-0"
			/>
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					size="sm"
					class="hidden gap-1.5 whitespace-nowrap
						text-xs sm:flex"
				>
					<svg
						xmlns="http://www.w3.org/2000/svg"
						width="14"
						height="14"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
						stroke-linecap="round"
						stroke-linejoin="round"
					>
						<rect
							width="5"
							height="5"
							x="3"
							y="3"
							rx="1"
						/>
						<rect
							width="5"
							height="5"
							x="16"
							y="3"
							rx="1"
						/>
						<rect
							width="5"
							height="5"
							x="3"
							y="16"
							rx="1"
						/>
						<rect
							width="5"
							height="5"
							x="16"
							y="16"
							rx="1"
						/>
					</svg>
					Scan QR
				</Button>
				<Button
					size="sm"
					class="whitespace-nowrap bg-foreground
						text-background hover:bg-foreground/90"
				>
					Find
				</Button>
			</div>
		</div>
	</div>
</section>
