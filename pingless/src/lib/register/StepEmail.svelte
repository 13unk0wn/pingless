<script>
	const apiUrl = import.meta.env.VITE_BACKEND_URL;
	import axios from 'axios';

	const { step } = $props();

	let email = $state();
	function sendEmail() {
		try {
			axios.post(`${apiUrl}/user/email_verification`, { email: email });
			step.update(n => n + 1)
		} catch (e) {
			console.log(e);
		}
	}
</script>

<form onsubmit={sendEmail}>
	<div class="w-full max-w-sm min-w-[200px]">
		<input
			class="ease w-full rounded-md border border-slate-200 bg-transparent px-3 py-2 text-sm text-slate-700 shadow-sm transition duration-300 placeholder:text-slate-400 hover:border-slate-300 focus:border-slate-400 focus:shadow focus:outline-none"
			placeholder="Enter your email"
			bind:value={email}
		/>
	</div>
	<button class="ease my-2 rounded-md border">Submit</button>
</form>
