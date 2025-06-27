<script>
  import StepEmail from "$lib/StepEmail.svelte"
	import { writable } from "svelte/store";
	const step = writable(1)
</script>



<div class="flex items-center justify-center min-h-screen">
  {#if $step === 1}
    <StepEmail {step}/>
  {/if}
</div>



