<script>
	/** @type {import('./$types').PageData} */
	export let data;

	$: ({ status, message, error } = data);
</script>

<main class="container">
	<h1>Eventio OS</h1>
	<p>Walking Skeleton Setup</p>

	<div class="status-card {status === 'ok' ? 'ok' : status === 'error' ? 'error' : ''}">
		<h2>Backend Status: {status}</h2>
		{#if message}
			<p>{message}</p>
		{/if}
		{#if error}
			<p class="error-msg">Error: {error}</p>
		{/if}
	</div>
</main>

<style>
	:global(body) {
		font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
		background: #f4f4f9;
		color: #333;
		margin: 0;
		padding: 2rem;
	}
	.container {
		max-width: 600px;
		margin: 0 auto;
	}
	h1 {
		color: #111;
	}
	.status-card {
		background: white;
		padding: 1.5rem;
		border-radius: 8px;
		box-shadow: 0 4px 6px rgba(0,0,0,0.1);
		margin-top: 2rem;
		border-left: 5px solid #ccc;
	}
	.status-card.ok {
		border-color: #28a745;
	}
	.status-card.error {
		border-color: #dc3545;
	}
	.error-msg {
		color: #dc3545;
		font-family: monospace;
	}
</style>
