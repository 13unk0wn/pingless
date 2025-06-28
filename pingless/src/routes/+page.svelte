<script>
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	let apiUrl = import.meta.env.VITE_BACKEND_URL;
	console.log(apiUrl);

	onMount(() => {
		let accessToken = localStorage.getItem('accessToken');
		if (accessToken == null) {
			goto('/login');
		}
	});
</script>

<h1>Main Page</h1>
