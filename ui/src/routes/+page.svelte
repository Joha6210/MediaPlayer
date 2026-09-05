<!-- src/routes/+page.svelte (Din nye start/loading-side) -->
<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation'; // SvelteKits indbyggede funktion til at skifte side
  import { getState } from '$lib/api.svelte';

  onMount(async () => {
    try {
      // Vent på at dit API svarer og data er hentet
      await getState(); 
      
      // Send brugeren videre til afspilleren automatisk
      goto('/player'); 
    } catch (err) {
      console.error("API ikke klar endnu", err);
    }
  });
</script>

<main>
  <p>Loading MediaPlayer...</p>
</main>