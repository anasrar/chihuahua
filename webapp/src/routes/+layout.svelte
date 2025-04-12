<script lang="ts">
	import "inter-ui/inter.css";
	import "../app.css";
	import * as Sidebar from "@/components/ui/sidebar";
	import { AppSidebar } from "@/components/app-sidebar";
	import { ModeWatcher } from "mode-watcher";
	import { Toaster } from "$lib/components/ui/sonner";

	let { children } = $props();
</script>

<ModeWatcher />
<Toaster position="top-right" expand={true} />
<Sidebar.Provider>
	<AppSidebar />
	<main class="relative w-full font-inter">
		{@render children?.()}
	</main>
</Sidebar.Provider>
