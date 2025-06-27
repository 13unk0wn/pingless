<script>
	import {StepEmail} from "$lib/register/StepEmail.svelte"
	let currentStep = $state(0);

	const steps = [StepEmail,StepOtp,StepUsername, StepPassword, StepConfirm];

	function prevStep() {
		currentStep = Math.max(currentStep - 1, 0);
	}
	function nextStep() {
		currentStep = Math.min(currentStep + 1, steps.length - 1);
	}
</script>


{@render steps[currentStep]()}

