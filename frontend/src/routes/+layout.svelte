<script lang="ts">
	import './layout.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href="/open_pass_logo.svg" />
</svelte:head>

{@render children()}
