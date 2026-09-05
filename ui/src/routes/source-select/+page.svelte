<script>
  import { getState, openStateSocket, selectSource, getSources} from '$lib/api.svelte';
  import { onDestroy, onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { clock } from '$lib/clock';

  let state = $state({
    activeSource: '',
    label: '',
    playing: false,
    volume: 60
  });
  let error = $state('');
  let socket;
  
  let sources = $state([]);
  let currentIndex = $state(0);

  async function refreshState() {
    state = await getState();
  }

  async function loadSources() {
    error = '';
    try {
      const res = await getSources();
      sources = (res.sources || []).map((item) => {
        const id = typeof item === 'object' ? item.name : item;
        return {
          id: id,
          name: id.toUpperCase(),
          data: item
        };
      });
    } catch (err) {
      error = err.message;
    }
  }

  // Handle confirming and switching to the currently focused source
  async function confirmSelection() {
    error = '';
    const selected = sources[currentIndex];
    if (!selected) return;

    try {
      if (selected.id === 'bluetooth') {
        state = await selectSource({ source: 'bluetooth' });
      } else if (selected.id === 'dr-radio') {
        const defaultStation = selected.data?.stations?.[3] || 'dr-p3';
        state = await selectSource({ source: 'dr-radio', station: defaultStation } );
      } else if (selected.id === 'plexamp') {
        state = await selectSource({ source: 'plexamp', meta: { path: '/audio/:/transcode/universal/start.m3u8' } });
      } else if (selected.id === 'internet-radio') {
        state = await selectSource({ source: 'internet-radio', url: 'https://example.com/stream.mp3' });
      } else {
        state = await selectSource({ source: selected.id });
      }
      goto('/'); // Navigate back to player screen on success
    } catch (err) {
      error = err.message;
    }
  }

    async function switchInternet() {
    error = '';
    try {
      state = await selectSource({
        source: 'internet-radio',
        title: 'Internet Radio',
        url: radioUrl
      });
    } catch (err) {
      error = err.message;
    }
  }

  

  function nextSource() {
    if (sources.length === 0) return;
    currentIndex = (currentIndex + 1) % sources.length;
  }

  function prevSource() {
    if (sources.length === 0) return;
    currentIndex = (currentIndex - 1 + sources.length) % sources.length;
  }

  function openSettings() {
    // Handle settings navigation or modal
  }

  onMount(async () => {
    await refreshState();
    await loadSources();
    socket = openStateSocket((nextState) => {
      state = nextState;
    });
  });

  onDestroy(() => {
    if (socket) {
      socket.close();
    }
  });

  // Calculates exact translation so the active card's center matches the viewport center
  // Inactive card = 160px width, Active card = 200px width, Gap = 32px
  let trackOffset = $derived((() => {
    const inactiveWidth = 160;
    const activeWidth = 200;
    const gap = 32;
    
    let offset = 0;
    for (let i = 0; i < currentIndex; i++) {
      offset += inactiveWidth + gap;
    }
    offset += activeWidth / 2;
    return offset;
  })());

  let dateStr = $derived(`${String($clock.getDate()).padStart(2, '0')}-${String($clock.getMonth() + 1).padStart(2, '0')}`);
  let timeStr = $derived($clock.toLocaleTimeString("da-DK", { 
    hour: '2-digit', 
    minute: '2-digit',
    timeZone: 'Europe/Copenhagen' 
  }));
</script>

<div class="player-background-wrapper">
  <main class="player-flex">

    <!-- Top Bar -->
    <header class="top-bar">
      <div class="spacer"></div>
      <div class="clock-display">
        <span>{dateStr}<br>{timeStr}</span>
      </div>
      <div style="display: flex; gap: 16px;">
        <button class="button settings-btn" onclick={openSettings} aria-label="Settings">
          <img src="/icons/settings.svg" alt="Settings" width="20" height="20" />
        </button>
      </div>
    </header>

    <!-- Main Carousel Selection Stage -->
    <section class="carousel-stage">
      

      <div class="carousel-viewport">
        <div 
          class="carousel-track"
          style="transform: translateX(calc(50vw - 20px - {trackOffset}px));"
        >
          {#each sources as source, index}
            <div 
              class="carousel-card"
              class:active={index === currentIndex}
            >
              {#if source.id === 'dr-radio'}
                <div class="icon-card-content">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 800" class="card-svg-fill">
                    <path d="M0 0v800h800V0z" fill="#ff001e"/>
                    <path d="M0 560h800v240H0z" fill="#1e1e1e"/>
                    <path fill="#fff" d="M319.27 618.67h-171.2a2.07 2.07 0 0 0-2.28 2.27v115.3a2.09 2.09 0 0 0 2.28 2.29h171.2c50.29 0 75.67-16.71 75.67-60.31 0-43.3-25.38-59.55-75.67-59.55zm-44.33 97.83h-53.63c-1.52 0-1.83-.61-1.83-1.83v-72.16c0-1.2.31-1.82 1.83-1.82h53.63c31.89 0 44.66 9.11 44.66 37.83s-12.77 37.98-44.66 37.98zM665 733.67l-49.52-34.94c-1.22-.92-2-1.37-2-2s.46-.92 1.53-.92c25 0 44.5-11.69 44.5-37.51s-16.65-39.63-47.51-39.63H424.4a2.06 2.06 0 0 0-2.27 2.27v115.3a2.08 2.08 0 0 0 2.27 2.29h69a2.08 2.08 0 0 0 2.28-2.29V708.9c0-1.37.46-1.82 1.83-1.82h39.73c2 0 2.73.15 4.1 1.22l34.49 28.56a6.6 6.6 0 0 0 4.85 1.67h85.84q1.83 0 1.83-1.38c.04-1.21-2.09-2.57-3.35-3.48zm-107.54-48h-59.81c-1.37 0-1.83-.46-1.83-1.82v-41.34c0-1.36.46-1.82 1.83-1.82h59.85c20.06 0 28.56 5.48 28.56 22 0 16.74-8.5 22.97-28.56 22.97zM233 371.64a1.79 1.79 0 0 1-1.78 1.8h-45.15a1.79 1.79 0 0 1-1.78-1.8V118.87a1.79 1.79 0 0 1 1.78-1.8h45.13a1.79 1.79 0 0 1 1.78 1.8zm75.55 0a1.79 1.79 0 0 1-1.78 1.8h-45.13a1.79 1.79 0 0 1-1.78-1.8v-81.88a1.79 1.79 0 0 1 1.78-1.8h45.13a1.79 1.79 0 0 1 1.78 1.8zm75.57-21.36a1.79 1.79 0 0 1-1.78 1.8h-45.13a1.79 1.79 0 0 1-1.77-1.8V204.39a1.79 1.79 0 0 1 1.77-1.8h45.13a1.79 1.79 0 0 1 1.78 1.8zm75.57 106.85a1.79 1.79 0 0 1-1.78 1.8h-45.12a1.79 1.79 0 0 1-1.78-1.8V204.36a1.79 1.79 0 0 1 1.78-1.8h45.12a1.79 1.79 0 0 1 1.78 1.8zM535.26 393a1.79 1.79 0 0 1-1.78 1.8h-45.12a1.79 1.79 0 0 1-1.78-1.8V247.1a1.79 1.79 0 0 1 1.78-1.8h45.12a1.79 1.79 0 0 1 1.78 1.8zm75.58 0a1.8 1.8 0 0 1-1.78 1.8h-45.13a1.79 1.79 0 0 1-1.78-1.8V140.22a1.79 1.79 0 0 1 1.78-1.8h45.13a1.8 1.8 0 0 1 1.78 1.8z"/>
                  </svg>
                </div>
              {:else if source.id === 'bluetooth'}
                <div class="icon-card-content">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 976" class="card-svg-fill bluetooth-svg">
                    <rect ry="291" height="976" width="640" fill="#0a3d91"/>
                    <path d="m157 330 305 307-147 178V179l147 170-305 299" stroke="#FFF" stroke-width="53" fill="none"/>
                  </svg>
                </div>
              {:else}
                <div class="icon-card-content">
                  <span class="text-normal" style="color: white; font-size: 16px;">{source.name}</span>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    </section>
    
    {#if error}
    <section class="floating-section error-text">
        <span>{error}</span>
    </section>
    {/if}

    <!-- Bottom Navigation Bar -->
    <footer class="floating-section bottom-bar">
      <section class="selection-controls">
        <button class="button nav-btn" onclick={prevSource} aria-label="Previous">
          <img src="/icons/back-play.svg" alt="Prev" width="20" height="20" />
        </button>
        <button class="button nav-btn confirm-btn" onclick={confirmSelection} aria-label="Confirm">
          <img src="/icons/pause.svg" alt="Confirm" width="20" height="20" />
        </button>
        <button class="button nav-btn" onclick={nextSource} aria-label="Next">
          <img src="/icons/forward.svg" alt="Next" width="20" height="20" />
        </button>
      </section>
    </footer>

  </main>
</div>

<style>
  .card-svg-fill {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .player-background-wrapper {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-size: cover;
    background-position: center;
    background-repeat: no-repeat;
    box-sizing: border-box;
    overflow: hidden;
    background-image: url('/img/background_image.jpg');
  }

  .player-flex {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    height: 100vh;
    padding: 20px;
    box-sizing: border-box;
  }

  .top-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }

  .clock-display {
    font-family: 'Geist', sans-serif;
    font-weight: 600;
    font-size: 13px;
    color: #A2A2AD;
    text-align: center;
    line-height: 1.2;
  }

  .carousel-stage {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    flex: 1;
    width: 100%;
    overflow: hidden;
  }

  .carousel-viewport {
    width: 100%;
    display: flex;
    justify-content: flex-start;
    overflow: hidden;
    padding: 30px 0;
    position: relative;
  }

  .carousel-track {
    display: flex;
    align-items: center;
    gap: 32px;
    transition: transform 0.4s cubic-bezier(0.25, 1, 0.5, 1);
    will-change: transform;
  }

  .carousel-card {
    width: 160px;
    height: 160px;
    border-radius: 24px;
    background: rgba(20, 20, 25, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.08);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.4s cubic-bezier(0.25, 1, 0.5, 1);
    opacity: 0.4;
    transform: scale(0.85);
    overflow: hidden;
  }

  .carousel-card.active {
    width: 200px;
    height: 200px;
    opacity: 1;
    transform: scale(1);
    box-shadow: 0 0 40px rgba(0, 240, 255, 0.25);
    border: 1px solid rgba(0, 240, 255, 0.4);
    background: transparent;
  }

  .icon-card-content {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
  }

  .floating-section {
    color: white;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border-radius: 24px;
    margin-inline: auto;
    width: fit-content;
    height: fit-content;
  }

  .bottom-bar {
    padding: 12px 24px;
    margin-bottom: 10px;
  }

  .selection-controls {
    display: flex;
    gap: 16px;
    align-items: center;
    justify-content: center;
  }

  .button {
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    background: rgba(255, 255, 255, 0.123);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 50%;
    width: 44px;
    height: 44px;
    cursor: pointer;
  }

  .confirm-btn {
    background: rgba(0, 240, 255, 0.15);
    border-color: rgba(0, 240, 255, 0.4);
  }

  .settings-btn {
    border-radius: 50%;
    width: 40px;
    height: 40px;
  }

  .bottom-bar img, .top-bar img {
    filter: brightness(0) invert(1);
  }

  .confirm-btn img {
    filter: none;
  }

  .error-text {
    color: #ff4d4d;
    font-family: 'Geist', sans-serif;
    font-size: 14px;
    margin-bottom: 10px;
    padding: 10px;
  }
</style>