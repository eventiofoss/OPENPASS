<script lang="ts">
</script>

<section class="px-8 py-14 sm:px-10 lg:px-12">
	<div class="mx-auto max-w-7xl">
		<h2 class="font-serif text-3xl italic sm:text-4xl">
			Explore On The Map
		</h2>
		<div
			class="mt-8 overflow-hidden border border-border
				bg-card"
		>
			<!-- Map placeholder -->
			<div
				class="relative flex h-[480px] items-center
					justify-center bg-gradient-to-br
					from-green-100/60 via-sky-100/40
					to-emerald-100/50 sm:h-[540px]"
			>
				<!-- Grid lines -->
				<svg
					class="absolute inset-0 h-full w-full opacity-20"
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
					<rect width="100%" height="100%" fill="url(#map-grid)" />
				</svg>

				<!-- Center marker -->
				<div class="relative z-10 text-center">
					<div class="flex flex-col items-center">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="44"
							height="44"
							viewBox="0 0 24 24"
							fill="#3D6B8C"
							stroke="#3D6B8C"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
						>
							<path
								d="M20 10c0 4.993-5.539 10.193-7.399 11.799a1 1 0 0 1-1.202 0C9.539 20.193 4 14.993 4 10a8 8 0 0 1 16 0"
							/>
							<circle cx="12" cy="10" r="3" fill="white" />
						</svg>
						<span
							class="mt-3 bg-white/90 px-4 py-2
								text-sm font-semibold text-foreground
								shadow-sm backdrop-blur-sm"
						>
							Karukutty, IN
						</span>
					</div>
				</div>

				<!-- Floating dots -->
				<div
					class="absolute left-[20%] top-[30%]
						h-4 w-4 animate-pulse bg-accent/60"
				></div>
				<div
					class="absolute right-[25%] top-[25%]
						h-3 w-3 animate-pulse bg-pink-400/60"
					style="animation-delay: 0.5s;"
				></div>
				<div
					class="absolute bottom-[35%] left-[35%]
						h-3.5 w-3.5 animate-pulse bg-amber-400/60"
					style="animation-delay: 1s;"
				></div>
				<div
					class="absolute bottom-[25%] right-[30%]
						h-3 w-3 animate-pulse bg-emerald-400/60"
					style="animation-delay: 1.5s;"
				></div>

				<!-- Zoom controls -->
				<div
					class="absolute bottom-4 right-4 z-10
						flex flex-col gap-1"
				>
					<button
						class="flex h-10 w-10 items-center
							justify-center border border-border
							bg-white text-foreground shadow-md
							transition-colors hover:bg-gray-50
							text-lg font-medium"
						aria-label="Zoom in"
					>
						+
					</button>
					<button
						class="flex h-10 w-10 items-center
							justify-center border border-border
							bg-white text-foreground shadow-md
							transition-colors hover:bg-gray-50
							text-lg font-medium"
						aria-label="Zoom out"
					>
						−
					</button>
				</div>
			</div>
		</div>
	</div>
</section>
