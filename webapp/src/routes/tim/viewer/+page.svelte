<script lang="ts">
	import * as Sidebar from "@/components/ui/sidebar";
	import { DropZoneOverlay } from "@/components/dropzone-overlay";
	import { Button } from "@/components/ui/button";
	import { FileIcon, InfoIcon, ImageIcon, ViewIcon, ChevronsUpDownIcon } from "@lucide/svelte";
	import * as Tooltip from "@/components/ui/tooltip";
	import { buttonVariants } from "@/components/ui/button";
	import { DataViewExt } from "@/dataviewerext";
	import { Tim2, timImageTypeToString, type Tim2Picture } from "@/formats/tim2";
	import { Tim3 } from "@/formats/tim3";
	import { Tm3 } from "@/formats/tm3";
	import { toast } from "svelte-sonner";
	import * as Panzoom from "@panzoom/panzoom";
	import { ScrollArea } from "$lib/components/ui/scroll-area";
	import * as Dialog from "$lib/components/ui/dialog";
	import * as Drawer from "$lib/components/ui/drawer";
	import * as Collapsible from "$lib/components/ui/collapsible";
	import { MediaQuery } from "svelte/reactivity";

	type Item = {
		filename: string;
		picture: Tim2Picture;
		data: ImageData;
		thumb: string;
	};

	let inputFile = $state<HTMLInputElement>();

	let canvas = $state<HTMLCanvasElement>();
	let panzoom = $state<Panzoom.PanzoomObject>();

	let items = $state<Item[]>([]);
	let currentItem = $state(0);

	const isDesktop = new MediaQuery("(min-width: 550px)");

	let isOpenInformation = $state(false);

	$effect(() => {
		panzoom = Panzoom.default(canvas!.parentElement!, {
			canvas: true,
		});

		canvas?.parentElement?.parentElement?.addEventListener("wheel", (ev) => {
			panzoom?.zoomWithWheel(ev);
		});
	});

	const ondrop = (files: FileList | null) => {
		if (files === null) {
			return;
		}

		const fs = Array.from(files);

		const r = new FileReader();
		r.addEventListener("load", () => {
			const buffer = r.result as ArrayBuffer;
			const decoder = new TextDecoder();
			const str = decoder.decode(buffer);

			const signatures = str.match(/TIM(2|3)/g) ?? [];
			if (signatures.length === 0 || !signatures[0]?.startsWith("TIM")) {
				const signatureError = new Error("signature for TIM not found");
				toast.error(signatureError.name, {
					description: signatureError.message,
					action: {
						label: "Close",
						onClick: () => console.log(signatureError),
					},
				});
				return;
			}

			const signature = signatures[0];
			const dv = new DataViewExt(buffer);

			switch (signature) {
				case "TIM2":
					const tim = new Tim2();

					const { error: parseError } = tim.fromDataViewExt(dv);
					if (parseError !== undefined) {
						toast.error(parseError.name, {
							description: parseError.message,
							action: {
								label: "Close",
								onClick: () => console.log(parseError),
							},
						});
						return;
					}

					const { value: image, error: imageError } = tim.pictureToImageData(0);
					if (imageError !== undefined) {
						toast.error(imageError.name, {
							description: imageError.message,
							action: {
								label: "Close",
								onClick: () => console.log(imageError),
							},
						});
						return;
					}

					canvas!.width = image.width;
					canvas!.height = image.height;

					const ctx = canvas!.getContext("2d")!;
					ctx.clearRect(0, 0, image.width, image.height);
					ctx.putImageData(image, 0, 0);

					const thumb = canvas!.toDataURL("image/png");

					items = [
						{
							filename: fs[0].name.replace(/\.[^/.]+$/, ""),
							picture: tim.pictures[0],
							data: image,
							thumb: thumb,
						},
					];
					break;
				case "TIM3":
					if (signatures.length === 1) {
						const tim = new Tim3();

						const { error: parseError } = tim.fromDataViewExt(dv);
						if (parseError !== undefined) {
							toast.error(parseError.name, {
								description: parseError.message,
								action: {
									label: "Close",
									onClick: () => console.log(parseError),
								},
							});
							return;
						}

						const { value: image, error: imageError } = tim.pictureToImageData(0);
						if (imageError !== undefined) {
							toast.error(imageError.name, {
								description: imageError.message,
								action: {
									label: "Close",
									onClick: () => console.log(imageError),
								},
							});
							return;
						}

						canvas!.width = image.width;
						canvas!.height = image.height;

						const ctx = canvas!.getContext("2d")!;
						ctx.clearRect(0, 0, image.width, image.height);
						ctx.putImageData(image, 0, 0);

						const thumb = canvas!.toDataURL("image/png");

						items = [
							{
								filename: fs[0].name.replace(/\.[^/.]+$/, ""),
								picture: tim.pictures[0],
								data: image,
								thumb: thumb,
							},
						];
					} else {
						const tm3 = new Tm3();

						const { error: parseError } = tm3.fromDataViewExt(dv);
						if (parseError !== undefined) {
							toast.error(parseError.name, {
								description: parseError.message,
								action: {
									label: "Close",
									onClick: () => console.log(parseError),
								},
							});
							return;
						}

						const newItems: Item[] = [];

						for (const entry of tm3.entries) {
							const tim = new Tim3();

							const { error: parseError } = tim.fromDataViewExt(dv, { offset: entry.offset });
							if (parseError !== undefined) {
								toast.error(parseError.name, {
									description: parseError.message,
									action: {
										label: "Close",
										onClick: () => console.log(parseError),
									},
								});
								return;
							}

							const { value: image, error: imageError } = tim.pictureToImageData(0);
							if (imageError !== undefined) {
								toast.error(imageError.name, {
									description: imageError.message,
									action: {
										label: "Close",
										onClick: () => console.log(imageError),
									},
								});
								return;
							}

							canvas!.width = image.width;
							canvas!.height = image.height;

							const ctx = canvas!.getContext("2d")!;
							ctx.clearRect(0, 0, image.width, image.height);
							ctx.putImageData(image, 0, 0);

							const thumb = canvas!.toDataURL("image/png");

							newItems.push({
								filename: entry.name,
								picture: tim.pictures[0],
								data: image,
								thumb: thumb,
							});
						}

						items = newItems;

						canvas!.width = items[0].data.width;
						canvas!.height = items[0].data.height;

						const ctx = canvas!.getContext("2d")!;
						ctx.clearRect(0, 0, items[0].data.width, items[0].data.height);
						ctx.putImageData(items[0].data, 0, 0);
					}
					break;
				default:
					const versionError = new Error("unknown TIM version");
					toast.error(versionError.name, {
						description: versionError.message,
						action: {
							label: "Close",
							onClick: () => console.log(versionError),
						},
					});
					return;
			}

			currentItem = 0;
		});
		r.readAsArrayBuffer(fs[0]);
	};

	const downloadAsPng = () => {
		if (items.length === 0) {
			return;
		}

		const url = canvas!.toDataURL("image/png");
		const link = document.createElement("a");
		link.href = url;
		link.download = `${items[currentItem].filename}.png`;
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
	};
</script>

<svelte:head>
	<title>TIM Viewer</title>
</svelte:head>

<!-- canvas -->
<div class="absolute inset-0 flex flex-col items-start justify-start">
	<div
		class="m-auto"
		class:opacity-0={items.length === 0}
		style="background: repeating-conic-gradient(#cccccc 0% 25%, #f0f0f0 0% 50%) 50% / 24px 24px"
	>
		<canvas bind:this={canvas}></canvas>
	</div>
</div>

<DropZoneOverlay {ondrop} hide={items.length !== 0} />

<!-- top left -->
<div class="absolute left-4 top-4">
	<div class="flex items-center gap-2">
		<Sidebar.Trigger />
		<div class="text-sm font-semibold tracking-tight">TIM Viewer</div>
	</div>
</div>

<!-- bottom -->
<div class="absolute bottom-0 left-0 right-0 flex flex-col">
	<div class="flex flex-row gap-2 px-4 pb-4">
		<input
			bind:this={inputFile}
			type="file"
			class="hidden"
			onchange={(ev) => {
				ondrop(ev.currentTarget.files);
			}}
		/>
		<Tooltip.Provider>
			<Tooltip.Root>
				<Tooltip.Trigger
					class={buttonVariants({ size: "icon" })}
					onclick={() => {
						inputFile?.click();
					}}
				>
					<FileIcon class="size-4" />
					<span class="sr-only">Browse File</span>
				</Tooltip.Trigger>
				<Tooltip.Content>
					<p>Browse File</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</Tooltip.Provider>
		<Tooltip.Provider>
			<Tooltip.Root>
				<Tooltip.Trigger
					class={buttonVariants({ size: "icon" })}
					onclick={() => {
						panzoom?.reset();
					}}
				>
					<ViewIcon class="size-4" />
					<span class="sr-only">Reset View</span>
				</Tooltip.Trigger>
				<Tooltip.Content>
					<p>Reset View</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</Tooltip.Provider>
		<Tooltip.Provider disabled={items.length === 0}>
			<Tooltip.Root>
				<Tooltip.Trigger
					class={buttonVariants({ size: "icon" })}
					disabled={items.length === 0}
					onclick={() => {
						isOpenInformation = !isOpenInformation;
					}}
				>
					<InfoIcon class="size-4" />
					<span class="sr-only">Information</span>
				</Tooltip.Trigger>
				<Tooltip.Content>
					<p>Information</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</Tooltip.Provider>
		<Button disabled={items.length === 0} onclick={downloadAsPng} class="ml-auto">
			<ImageIcon class="mr-2 size-4" />
			Download as PNG
		</Button>
	</div>
	{#if items.length > 1}
		<div>
			<ScrollArea class="whitespace-nowrap" orientation="horizontal">
				<div class="flex w-max space-x-4 px-4 pb-4">
					{#each items as item, index}
						<button
							onclick={() => {
								canvas!.width = items[index].data.width;
								canvas!.height = items[index].data.height;

								const ctx = canvas!.getContext("2d")!;
								ctx.clearRect(0, 0, items[index].data.width, items[index].data.height);
								ctx.putImageData(items[index].data, 0, 0);

								currentItem = index;
							}}
						>
							<figure class="shrink-0">
								<div
									class="overflow-hidden rounded-md"
									style="background: repeating-conic-gradient(#cccccc 0% 25%, #f0f0f0 0% 50%) 50% / 24px 24px"
								>
									<img
										src={item.thumb}
										alt={item.filename}
										class="pointer-events-none aspect-square h-20 w-20 object-cover"
									/>
								</div>
								<figcaption
									class="select-none pt-2 text-center text-xs font-semibold text-foreground"
								>
									{item.filename}
								</figcaption>
							</figure>
						</button>
					{/each}
				</div>
			</ScrollArea>
		</div>
	{/if}
</div>
{#if items.length !== 0}
	{#if isDesktop.current}
		<Dialog.Root bind:open={isOpenInformation}>
			<Dialog.Content class="sm:max-w-[480px]">
				<div class="flex flex-col gap-4">
					<div class="grid grid-cols-2 gap-1">
						<div class="flex flex-col gap-1">
							<div class="text-sm font-semibold tracking-tight">Dimensions</div>
							<div class="text-xs font-medium text-muted-foreground">
								{items[currentItem].picture.imageWidth} x {items[currentItem].picture.imageHeight}
							</div>
						</div>
						<div class="flex flex-col gap-1">
							<div class="text-sm font-semibold tracking-tight">Type</div>
							<div class="text-xs font-medium text-muted-foreground">
								{timImageTypeToString(items[currentItem].picture.imageType)}
							</div>
						</div>
					</div>
					<div class="flex flex-col gap-1">
						<Collapsible.Root open={true} class="flex flex-col gap-1">
							<div class="flex flex-row gap-1">
								<div class="flex flex-col gap-1">
									<div class="text-sm font-semibold tracking-tight">CLUT Colors</div>
									<div class="text-xs font-medium text-muted-foreground">
										{items[currentItem].picture.clutColors}
									</div>
								</div>
								<div class="ml-auto">
									<Collapsible.Trigger class={buttonVariants({ size: "icon", variant: "outline" })}>
										<ChevronsUpDownIcon class="size-4" />
									</Collapsible.Trigger>
								</div>
							</div>
							<Collapsible.Content>
								<div class="grid grid-cols-[repeat(16,_1fr)]">
									{#each items[currentItem].picture.clutData as clut}
										<div
											style={`aspect-ratio: 1/1; background: #${clut[0].toString(16).padStart(2, "0")}${clut[1].toString(16).padStart(2, "0")}${clut[2].toString(16).padStart(2, "0")}${clut[3].toString(16).padStart(2, "0")}`}
										></div>
									{/each}
								</div>
							</Collapsible.Content>
						</Collapsible.Root>
					</div>
				</div>
			</Dialog.Content>
		</Dialog.Root>
	{:else}
		<Drawer.Root bind:open={isOpenInformation}>
			<Drawer.Content>
				<div class="flex flex-col gap-4 px-4 pt-2">
					<div class="grid grid-cols-2 gap-1">
						<div class="flex flex-col gap-1">
							<div class="text-sm font-semibold tracking-tight">Dimensions</div>
							<div class="text-xs font-medium text-muted-foreground">
								{items[currentItem].picture.imageWidth} x {items[currentItem].picture.imageHeight}
							</div>
						</div>
						<div class="flex flex-col gap-1">
							<div class="text-sm font-semibold tracking-tight">Type</div>
							<div class="text-xs font-medium text-muted-foreground">
								{timImageTypeToString(items[currentItem].picture.imageType)}
							</div>
						</div>
					</div>
					<div class="flex flex-col gap-1">
						<Collapsible.Root open={true} class="flex flex-col gap-1">
							<div class="flex flex-row gap-1">
								<div class="flex flex-col gap-1">
									<div class="text-sm font-semibold tracking-tight">CLUT Colors</div>
									<div class="text-xs font-medium text-muted-foreground">
										{items[currentItem].picture.clutColors}
									</div>
								</div>
								<div class="ml-auto">
									<Collapsible.Trigger class={buttonVariants({ size: "icon", variant: "outline" })}>
										<ChevronsUpDownIcon class="size-4" />
									</Collapsible.Trigger>
								</div>
							</div>
							<Collapsible.Content>
								<div class="grid grid-cols-[repeat(16,_1fr)]">
									{#each items[currentItem].picture.clutData as clut}
										<div
											style={`aspect-ratio: 1/1; background: #${clut[0].toString(16).padStart(2, "0")}${clut[1].toString(16).padStart(2, "0")}${clut[2].toString(16).padStart(2, "0")}${clut[3].toString(16).padStart(2, "0")}`}
										></div>
									{/each}
								</div>
							</Collapsible.Content>
						</Collapsible.Root>
					</div>
				</div>
				<Drawer.Footer>
					<Drawer.Close class={buttonVariants({ variant: "outline" })}>Close</Drawer.Close>
				</Drawer.Footer>
			</Drawer.Content>
		</Drawer.Root>
	{/if}
{/if}
