<script lang="ts">
	import { Button } from "$lib/components/ui/button/index.js";
	import { Input } from "$lib/components/ui/input/index.js";
</script>

<section class="px-8 pb-8 pt-12 sm:px-10 lg:px-12">
	<div class="mx-auto w-full max-w-7xl">
		<!-- Heading -->
		<h1
			class="font-serif text-6xl leading-tight
				tracking-tight sm:text-7xl lg:text-8xl"
		>
			<span class="font-normal italic">Find</span>
			<span class="font-normal text-[#8A8A8A]">
				Whats Happening
			</span>
		</h1>

		<!-- Location -->
		<div class="mt-2 flex items-center gap-2">
			<span class="text-lg text-muted-foreground">
				Near You,
			</span>
			<span class="text-lg font-semibold text-foreground">
				Karukutty, IN
			</span>
			<button
				class="ml-1 text-accent transition-colors
					hover:text-accent/80"
				aria-label="Change location"
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
				>
					<path
						d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"
					/>
					<path d="m15 5 4 4" />
				</svg>
			</button>
		</div>

		<!-- Search Bar — rectangle, left-aligned with heading -->
		<div
			class="mt-8 w-full
				border border-border bg-card shadow-sm
				flex items-center gap-3 pl-6 pr-2 py-2.5"
		>
			<!-- Search icon -->
			<svg
				xmlns="http://www.w3.org/2000/svg"
				width="22"
				height="22"
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
			<!-- Input -->
			<Input
				type="text"
				placeholder="Search Events, Organizers, or Topics..."
				class="flex-1 !border-0 !bg-transparent
					!shadow-none !outline-none !ring-0
					!text-base focus-visible:!ring-0"
			/>
			<!-- Scan QR -->
			<Button
				variant="outline"
				size="sm"
				class="hidden gap-2 whitespace-nowrap
					!rounded-none text-sm sm:flex"
			>
				<img
					src="/scan-qr-code.svg"
					alt="Scan QR"
					class="h-4 w-4"
				/>
				Scan QR
			</Button>
			<!-- Find -->
			<Button
				size="sm"
				class="!rounded-none whitespace-nowrap
					bg-foreground px-8 py-5 text-sm
					text-background hover:bg-foreground/90"
			>
				Find
			</Button>
		</div>
	</div>
</section>
