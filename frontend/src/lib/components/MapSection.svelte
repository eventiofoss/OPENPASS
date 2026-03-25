<script lang="ts">
</script>

<section class="px-4 py-8 sm:px-6">
	<div class="mx-auto max-w-6xl">
		<h2
			class="font-serif text-xl italic sm:text-2xl"
		>
			Explore On The Map
		</h2>
		<div
			class="mt-4 overflow-hidden rounded-xl
				border border-border bg-card"
		>
			<!-- Map placeholder -->
			<div
				class="relative flex h-[280px] items-center
					justify-center bg-gradient-to-br
					from-green-100/60 via-sky-100/40
					to-emerald-100/50 sm:h-[340px]"
			>
				<!-- Map-like grid lines -->
				<svg
					class="absolute inset-0 h-full w-full
						opacity-20"
					xmlns="http://www.w3.org/2000/svg"
				>
					<defs>
						<pattern
							id="map-grid"
							width="40"
							height="40"
							patternUnits="userSpaceOnUse"
						>
							<path
								d="M 40 0 L 0 0 0 40"
								fill="none"
								stroke="#888"
								stroke-width="0.5"
							/>
						</pattern>
					</defs>
					<rect
						width="100%"
						height="100%"
						fill="url(#map-grid)"
					/>
				</svg>

				<!-- Center marker -->
				<div class="relative z-10 text-center">
					<div class="flex flex-col items-center">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="32"
							height="32"
							viewBox="0 0 24 24"
							fill="#4E4EFE"
							stroke="#4E4EFE"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path
								d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0"
							/>
							<circle
								cx="12"
								cy="10"
								r="3"
								fill="white"
							/>
						</svg>
						<span
							class="mt-2 rounded-full
								bg-white/90 px-3 py-1
								text-xs font-medium
								text-foreground shadow-sm
								backdrop-blur-sm"
						>
							Karukutty, IN
						</span>
					</div>
				</div>

				<!-- Floating event dots -->
				<div
					class="absolute left-[20%] top-[30%]
						h-3 w-3 animate-pulse
						rounded-full bg-accent/60"
				></div>
				<div
					class="absolute right-[25%] top-[25%]
						h-2 w-2 animate-pulse
						rounded-full bg-pink-400/60"
					style="animation-delay: 0.5s;"
				></div>
				<div
					class="absolute bottom-[35%] left-[35%]
						h-2.5 w-2.5 animate-pulse
						rounded-full bg-amber-400/60"
					style="animation-delay: 1s;"
				></div>
				<div
					class="absolute bottom-[25%] right-[30%]
						h-2 w-2 animate-pulse
						rounded-full bg-emerald-400/60"
					style="animation-delay: 1.5s;"
				></div>

				<!-- Zoom controls -->
				<div
					class="absolute bottom-3 right-3 z-10
						flex flex-col gap-1"
				>
					<button
						class="flex h-8 w-8 items-center
							justify-center rounded bg-white
							text-foreground shadow-md
							transition-colors
							hover:bg-gray-50"
						aria-label="Zoom in"
					>
						+
					</button>
					<button
						class="flex h-8 w-8 items-center
							justify-center rounded bg-white
							text-foreground shadow-md
							transition-colors
							hover:bg-gray-50"
						aria-label="Zoom out"
					>
						−
					</button>
				</div>
			</div>
		</div>
	</div>
</section>
