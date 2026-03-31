<script lang="ts">
	const mapEmbedUrl =
		"https://www.openstreetmap.org/export/embed.html?bbox=76.35%2C10.18%2C76.42%2C10.24&layer=mapnik&marker=10.2126%2C76.3778";
</script>

<section class="mb-[60px]">
	<div class="w-full border border-[#D9D9D5] bg-white p-4 sm:p-6">
			<h2 class="font-serif text-[44px] leading-[0.92] font-normal tracking-[-0.01em] text-[#1E1E1E] sm:text-[48px]">
				Explore On The Map
			</h2>
			<div class="mt-6 overflow-hidden rounded-[2px] border border-[#D9D9D5]">
				<iframe
					title="Event map around Karukutty"
					src={mapEmbedUrl}
					class="h-[360px] w-full sm:h-[420px]"
					loading="lazy"
				></iframe>
			</div>
			<p class="mt-3 text-sm text-[#5A5A58]">
				Explore nearby event locations around Karukutty directly on the map.
			</p>
	</div>
</section>
