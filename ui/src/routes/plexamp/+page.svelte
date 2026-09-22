<script>
	import { goto } from '$app/navigation';

	const plexampUrl = 'http://127.0.0.1:32500';

	function openSourceSelect() {
		goto('/source-select');
	}
</script>

<svelte:head>
	<title>Plexamp</title>
</svelte:head>

<main class="plexamp-page">
	<header class="top-bar">
		<button class="button source-button" onclick={openSourceSelect} aria-label="Go to source selection">
			<img src="/icons/source-icon.svg" alt="" width="28" height="28" />
			<span>Source Select</span>
		</button>
		<h1>Plexamp</h1>
		<span class="status">Local server</span>
	</header>

	<section class="plexamp-frame" aria-label="Local Plexamp server">
		<iframe
			src={plexampUrl}
			title="Local Plexamp server"
			allow="autoplay; fullscreen"
		></iframe>
	</section>
</main>

<style>
	:global(body) {
		margin: 0;
		background: #111217;
		color: #ffffff;
		font-family: 'Geist', sans-serif;
	}

	.plexamp-page {
		min-height: 100vh;
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
		gap: 16px;
		padding: 16px;
		background: url('/img/background_image.jpg') center / cover fixed;
	}

	.top-bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
	}

	h1 {
		margin: 0;
		font-size: 24px;
	}

	.button {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		border: 1px solid rgba(255, 255, 255, 0.18);
		border-radius: 24px;
		padding: 10px 14px;
		color: #ffffff;
		background: rgba(255, 255, 255, 0.14);
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}

	.button:hover,
	.button:focus-visible {
		background: rgba(0, 240, 255, 0.24);
		outline: 2px solid #00f0ff;
		outline-offset: 2px;
	}

	.status {
		color: #a2a2ad;
		font-size: 13px;
	}

	.plexamp-frame {
		position: relative;
		flex: 1;
		min-height: 0;
		overflow: hidden;
		border: 1px solid rgba(255, 255, 255, 0.15);
		border-radius: 16px;
		background: #0c0d10;
	}

	iframe {
		display: block;
		width: 100%;
		height: calc(100vh - 88px);
		min-height: 520px;
		border: 0;
		background: #0c0d10;
	}

	@media (max-width: 600px) {
		.plexamp-page {
			padding: 10px;
		}

		.top-bar {
			flex-wrap: wrap;
		}

		h1 {
			order: -1;
			width: 100%;
			text-align: center;
		}

		.status {
			display: none;
		}

		iframe {
			height: calc(100vh - 78px);
			min-height: 0;
		}
	}
</style>
