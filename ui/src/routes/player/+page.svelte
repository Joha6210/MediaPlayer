<script >
  import { getState, openStateSocket, getStations, setVolume } from '$lib/api.svelte';
  import { onDestroy, onMount } from 'svelte';
  import { fly, fade } from 'svelte/transition';
  import { goto } from '$app/navigation';
  import { clock } from '$lib/clock';

  let state = $state({
    activeSource: '',
    label: '',
    playing: false,
    volume: 60,
    stream_title: '', // Modtager live-tekst ("Kunstner - Sang" eller kanalnavn)
    paused: false,     // Modtager sand/falsk pause-status fra MPV
    station: null, // Modtager den aktive station fra MPV
  }

);

  let stations = $state([]);

  let displayTitle = $derived(state.stream_title.replace('/', ' ') || state.label || 'Ingen stream kører');
  let displaySubtitle = $derived(state.station?.name ? state.station.name : 'Live Stream');

  let error = '';
  let plexPath = '/audio/:/transcode/universal/start.m3u8';
  let socket;
  let title = 'Midnight Drive';
  let artist = 'SYNTHWAVE';
  let album = 'Synthwave Collection';
  let format = $derived(
  (state.activeSource === 'dr-radio' || state.activeSource === 'INTERNET') && state.station 
    ? `${state.station?.codec || 'UNKNOWN'} • ${state.station?.bitrate ? state.station?.bitrate + ' KBPS' : ''}`.trim().replace(/•\s*$/, '')
    : 'FLAC 24-BIT 96KHZ'
); let progress = 40;
  let drStation = $derived(state.activeSource === 'DR' ? state.label : '');
  
	
  

  let dateStr = `${String($clock.getDate()).padStart(2, '0')}-${String($clock.getMonth() + 1).padStart(2, '0')}`;
  let timeStr = $clock.toLocaleTimeString("da-DK", { 
    hour: '2-digit', 
    minute: '2-digit',
    timeZone: 'Europe/Copenhagen' 
  });

  let { isOpen = $bindable(false), children } = $props();

  async function refreshState() {
    state = await getState();
  }

  async function onStationChange(event) {
    const selectedStationUUID = event.currentTarget.value;
    const selectedStation = Array.isArray(stations) ? stations.find(station => station.stationuuid === selectedStationUUID) : null;
    if (selectedStation) {
      console.log(`Selected station: ${selectedStation}`);
      try {
        state = await selectSource({ source: 'dr-radio', station: selectedStation});
      } catch (err) {
        error = err.message;
      }
    }
  }

  async function onVolumeChange(event) {
    const volume = Number(event.currentTarget.value);
    try {
      state = await setVolume(volume);
    } catch (err) {
      error = err.message;
    }
  }

  function sourceSelect() {
    goto('/source-select');
  }

  onMount(async () => {
    await refreshState();
    socket = openStateSocket((nextState) => {
      state = nextState;
    });

    if (state.activeSource === 'INTERNET' || state.activeSource === 'dr-radio') {
      stations = await getStations();
    }
    

  });

  onDestroy(() => {
    if (socket) {
      socket.close();
    }
  });

  

</script>

<div class="player-background-wrapper">
<main class="player-flex">

  <header class="top-bar">

    <button class="source-button button" onclick={sourceSelect} aria-label="Source Selection">
    <img src="/icons/source-icon.svg" alt="Source Icon" width="24" height="24" />
    <span class="text-normal source-text">
      <span class="source-label text-synth" >SOURCE</span>
      <span class="source-name">{state.activeSource}</span>
    </span>
    </button>

    <div class="clock-display">
      <span >{dateStr}<br>{timeStr}</span>
    </div>
    <div style="display: flex; gap: 16px;">
      <button class="button" onclick={() => {isOpen=true}} aria-label="Playlist"><img src="/icons/playlist.svg" alt="Playlist Icon" width="20" height="20" /></button>
      <button class="button" aria-label="Settings"><img src="/icons/settings.svg" alt="Settings Icon" width="20" height="20" /></button>
    </div>
  </header>

  <section class="main-stage">
    <div class="album">
      <div>
        <img src="/img/art_placeholder.png" alt="Album Cover" width="280" height="280" />
      </div>
    </div>
    
    <div class="track-info text-normal">
      <h1 style="margin-bottom: 0px;">{displayTitle}</h1>
      <h2 style="margin: 4px 0px 32px 0px;">{displaySubtitle} &#x2022; {artist}</h2>
      <div style="display: flex; flex-direction: row; justify-content: flex-start; gap: 8px; margin-bottom: 16px;">
        <span class="floating-section format-badge text-normal" style="width: fit-content; margin-right: 8px;">{format}</span>
      {#if state.activeSource === 'INTERNET' || state.activeSource === 'dr-radio'}
      <div>
        <select onchange={onStationChange} bind:value={stations} class="floating-section text-normal" style="width: 80%; padding: 8px 8px; border-radius: 20px; margin-inline: auto;">
          {#each stations as station (station.stationuuid)}
            <option class="text-normal" value={station.stationuuid}>{station.name}</option>
          {/each}
        </select>
      </div>
      {/if}
      </div>
    </div>
  </section>

  

  <footer class="floating-section bottom-bar" style="width:80%; ">
    <section class="top-row-bottom-bar">
      <section class="track-progress">
      <span class="text-normal">0:00</span>
      <input 
        type="range" 
        min="0" 
        max="100" 
        value="0" 
        style="--value: {progress}% "
      />
      <span class="text-normal">3:45</span>
    </section>
    </section>
      <section class="bottom-row-bottom-bar">
        <section class="shuffle-repeat">
          <button class="button" aria-label="Shuffle"><img src="/icons/shuffle.svg" alt="Shuffle Icon" width="20" height="20" /></button>
          <button class="button" aria-label="Repeat"><img src="/icons/repeat.svg" alt="Repeat Icon" width="20" height="20" /></button>
        </section>
        <section class="playback-controls">
          <button class="button" aria-label="Previous Track"><img src="/icons/back-play.svg" alt="Previous Icon" width="24" height="24" /></button>
          <button class="button play-button" aria-label={state.playing ? "Pause" : "Play"} onclick={() => state.playing = !state.playing}>
          <img src={state.playing ? "/icons/pause.svg" : "/icons/pause.svg"} alt={state.playing ? "Pause Icon" : "Play Icon"} width="32" height="32" /></button>
          <button class="button" aria-label="Next Track"><img src="/icons/forward.svg" alt="Next Icon" width="24" height="24" /></button>
        </section>
      <section class="volume-control">
        <img src="/icons/volume-lower.svg" alt="Volume Icon" width="20" height="20" />
        <input 
        type="range" 
        min="0" 
        max="100" 
        bind:value={state.volume} 
        onchange={onVolumeChange} 
        style="--value: {state.volume}%"/>
        <img src="/icons/volume-higher.svg" alt="Volume Icon" width="20" height="20" />
      </section>
    </section>
  </footer>


  <!-- SIDE PANEL OVERLAY INTEGRATION -->
    {#if isOpen}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div 
        class="backdrop" 
        transition:fade={{ duration: 200 }} 
        onclick={() => isOpen = false}
      ></div>

      <div 
        class="sidebar-panel" 
        transition:fly={{ x: 400, duration: 300 }}
      >
      <div class="sidebar-panel-top">
        <button class="button"  onclick={() => isOpen = false} aria-label="Close Sidebar">
          <svg stroke="white" stroke-width="1" fill="white" width="24" height="24"> <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z "/> </svg>
        </button>
        <h2 style="align-self: flex-start">Playlist</h2>
      </div>
        <div class="sidepanel-content">
          
        </div>
      </div>
    {/if}

</main>
</div>

<style>

  /* Base range styling */
  input[type="range"] {
    -webkit-appearance: none;
    appearance: none;
    background: transparent;
    border: none;
    outline: none;
    flex: 1;
    height: 10px;
    border-radius: 24px;
    /* Dynamically updates the split between white and background gray */
    background: linear-gradient(to right, white var(--value, 50%), rgba(255, 255, 255, 0.15) var(--value, 50%));
  }

  /* Chrome, Safari, Edge Track */
  input[type="range"]::-webkit-slider-runnable-track {
    height: 10px;
    border-radius: 24px;
    border: none;
    background: transparent;
  }

  /* Chrome, Safari, Edge Thumb (Hidden) */
  input[type="range"]::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 10px;
    height: 10px;
    visibility: hidden;
  }

  /* Firefox Track & Progress */
  input[type="range"]::-moz-range-track {
    height: 10px;
    border-radius: 24px;
    border: none;
    background: rgba(255, 255, 255, 0.15);
  }

  input[type="range"]::-moz-range-progress {
    background-color: white;
    height: 10px;
    border-radius: 24px;
  }

  /* Firefox Thumb (Hidden) */
  input[type="range"]::-moz-range-thumb {
    visibility: hidden;
    border: none;
  }

  .track-progress {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    width: 80%;
    flex:1; /* Adjust this to make the playback bar wider or narrower */
  }

  .top-row-bottom-bar{
    display: flex;
    justify-content: center;
    width: 100%;

  }


  .bottom-row-bottom-bar{
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    gap: inherit;
  }

  .volume-control {
  align-items: center;
  display: flex;
  gap: 8px;
  }

  .bottom-bar {
    align-items: center;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    padding: 1rem;
    border-radius: 24px; 
    gap: 16px;
    width: 100%;
  }

  .shuffle-repeat {
    display: flex;
    gap: 4px;
  }

  .button{
    display: flex;
    align-items: center;
    justify-content: center;
    color: white;
    background: rgba(255, 255, 255, 0.123);
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 8px 8px;
    border-radius: 28px;
    width: fit-content;
    height: fit-content;
  }

  .playback-controls {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .source-button {
    font-size: 14px;
    letter-spacing: 0.05em;
    display: grid;
    grid-template-columns: auto auto;
    align-items: center;
    gap: 8px;
    width: auto;
    height: 48px;
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

  .floating-section {
    color: white;
    background: rgba(255, 255, 255, 0.123);
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 8px 8px;
    width: fit-content;
    border-radius: 20px;
    margin-inline: auto
  }

  .player-flex {
    align-items: left;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding-bottom: 20px;
    padding-top: 5px;
  }

  .top-bar {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    padding: 1rem;
    border-radius: 24px;  
  }

  .main-stage {
    display: flex;
    flex-direction: row;
    justify-content: flex-start;
    padding: 1rem;
    border-radius: 24px;  
  }

  

  .source-text {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .track-info {
    display: flex;
    flex-direction: column;
    margin-right: auto;
    gap: 4px;
    margin-top: 16px;
  }

  .format-badge{
    margin-right: auto;
    margin-left: 0px;
  }

  h1 {
    font-family: 'Geist', sans-serif;
    color: #FFFFFF;
    font-weight: bold;
    font-size: 44px;
  }

  h2 {
    font-family: 'Geist', sans-serif;
    color: #A2A2AD;
    font-weight: bold;
    font-size: 22px;
  }

  .text-synth{
    font-family: 'Geist', sans-serif;
    font-weight: 600;
    font-size: 14px;
    color: #00F0FF;
    }
  .text-normal {
    font-family: 'Geist', sans-serif;
    font-weight: bold;
    font-size: 12px;
    color: #A2A2AD;
  }
  .clock-display {
    font-family: 'Geist', sans-serif;
    font-weight: 600;
    font-size: 14px;
    color: #A2A2AD;
    display: flex;
    align-items: center;
  }

  /* Dimmed full-screen backdrop background */
	.backdrop {
		position: fixed;
		inset: 0;
		background-color: rgba(0, 0, 0, 0.5);
		z-index: 40;
	}

	/* Panel container pushed to the right side */
	.sidebar-panel {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		width: 100%;
    max-height: 90%;
		max-width: 400px;
		background-color: white;
    background: rgba(255, 255, 255, 0.14);
		box-shadow: -4px 0 25px -5px rgba(0, 0, 0, 0.3);
		z-index: 50;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
    border-radius: 24px 0px 0px 24px;
	}

	.sidebar-panel-top {
    display: flex;
    flex-direction: row;
		justify-content: flex-start;
    align-items: center;
    gap: 8px;
		cursor: pointer;
		background: none;
		border: none;
		font-size: 1rem;
	}

	.sidepanel-content {
		margin-top: 1rem;
		flex-grow: 1;
		overflow-y: auto;
	}

</style>



