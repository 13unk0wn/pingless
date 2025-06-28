<script>
	let otp = "";
	import { InputOTP, InputOTPGroup, InputOTPSlot } from "$lib/components/ui/input-otp";
	import { Button } from '$lib/components/ui/button';

</script>

<form on:submit|preventDefault={() => console.log(otp)}>
	<div class="w-full max-w-sm min-w-[200px]">
		<InputOTP bind:value={otp} maxlength={6}>
			<InputOTPGroup>
				<InputOTPSlot />
				<InputOTPSlot />
				<InputOTPSlot />
				<InputOTPSlot />
				<InputOTPSlot />
				<InputOTPSlot />
			</InputOTPGroup>
		</InputOTP>
	</div>
	<Button class="ease my-2 rounded-md border">Submit</Button>
</form>
