<script>
  import StepEmail from "$lib/register/StepEmail.svelte"
  import StepOtp from "$lib/register/StepOtp.svelte" 
	import { writable } from "svelte/store";
	const step = writable(2)
</script>



<div class="flex items-center justify-center min-h-screen">
  {#if $step === 1}
    <StepEmail {step}/>
  {:else if $step === 2}
  <StepOtp/>
  {/if}
</div>



