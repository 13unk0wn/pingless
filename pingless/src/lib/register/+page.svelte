<script>
	import StepEmail from './StepEmail.svelte';
	import StepPassword from './StepPassword.svelte';

	let step = 1;

	function nextStep() {
		step = step + 1;
	}
</script>

{#if step === 1}
	<StepEmail on:next={nextStep} />
{:else if step === 2}
	<StepPassword />
{/if}
