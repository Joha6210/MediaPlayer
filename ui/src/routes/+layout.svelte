<script lang="ts">
  // Importer den globale CSS-fil her
  import '$lib/app.css';
  
  // Modtag de sider, som SvelteKit indlæser (f.eks. din afspiller)
  let { children } = $props(); // Svelte 5 syntax

</script>
<main> 
</main>

<!-- Render de forskellige undersider -->
{@render children()}