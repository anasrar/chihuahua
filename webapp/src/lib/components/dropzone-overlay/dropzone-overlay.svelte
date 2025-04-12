<script lang="ts">
	import { Button } from "@/components/ui/button";
	import { FileIcon } from "@lucide/svelte";

	let {
		multiple = false,
		hide = false,
		ondrop = () => {},
	}: {
		multiple?: boolean;
		hide?: boolean;
		ondrop?: (files: FileList | null) => void;
	} = $props();

	let inputFile = $state<HTMLInputElement>();
	let isDragOver = $state(false);

	$effect(() => {
		const dragover = () => {
			isDragOver = true;
		};
		const dragleave = () => {
			isDragOver = false;
		};

		document.addEventListener("dragover", dragover);
		document.addEventListener("dragleave", dragleave);

		return () => {
			document.removeEventListener("dragover", dragover);
			document.removeEventListener("dragleave", dragleave);
		};
	});
</script>

<div
	class="absolute inset-0 flex flex-col items-center justify-center gap-4"
	class:opacity-0={!isDragOver && hide}
	class:pointer-events-none={!isDragOver && hide}
	role="region"
	ondragover={(ev) => {
		ev.preventDefault();

		isDragOver = true;
	}}
	ondragleave={(ev) => {
		ev.preventDefault();

		isDragOver = false;
	}}
	ondrop={(ev) => {
		ev.preventDefault();

		isDragOver = false;

		ondrop(ev.dataTransfer?.files ?? null);
	}}
>
	{#if !isDragOver}
		<div class="select-none text-xl font-bold tracking-tight">
			Drag and Drop File{multiple ? "s" : ""} Here
		</div>
		<input
			bind:this={inputFile}
			type="file"
			class="hidden"
			{multiple}
			onchange={(ev) => {
				ondrop(ev.currentTarget.files);
			}}
		/>
		<div>
			<Button
				onclick={() => {
					inputFile?.click();
				}}
			>
				<FileIcon class="mr-2 size-4" />
				Browse File{multiple ? "s" : ""}
			</Button>
		</div>
	{:else}
		<div class="pointer-events-none select-none text-xl font-bold tracking-tight">
			Drop Here to Continue
		</div>
	{/if}
</div>
