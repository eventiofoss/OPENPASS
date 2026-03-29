<script lang="ts">
	import Navbar from "$lib/components/Navbar.svelte";
	import Footer from "$lib/components/Footer.svelte";
	import DashboardSidebar
		from "$lib/components/dashboard/DashboardSidebar.svelte";
	import { page } from "$app/state";
	import type { LayoutData } from "./$types";

	let { data, children } = $props<{
		data: LayoutData;
		children: import("svelte").Snippet;
	}>();

	let activeEventId = $derived(page.params.id ?? "");
</script>

<svelte:head>
	<title>Dashboard | Open Pass</title>
	<meta
		name="description"
		content="Organizer analytics dashboard for your
		events on Open Pass."
	/>
</svelte:head>

<div class="flex min-h-screen flex-col">
	<Navbar />

	<div
		class="mx-auto flex w-full max-w-[90rem] flex-1
			px-4 py-6 sm:px-6 lg:px-8"
	>
		<div
			class="grid w-full grid-cols-1
				lg:grid-cols-[16rem_1fr]"
		>
			<!-- Sidebar -->
			<div class="hidden lg:block">
				<DashboardSidebar
					events={data.events}
					{activeEventId}
				/>
			</div>

			<!-- Main content -->
			<main class="min-w-0">
				{@render children()}
			</main>
		</div>
	</div>

	<Footer />
</div>
