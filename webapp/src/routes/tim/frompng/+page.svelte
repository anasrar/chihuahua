<script lang="ts">
	import * as Sidebar from "@/components/ui/sidebar";
	import { DropZoneOverlay } from "@/components/dropzone-overlay";
	import { Button } from "@/components/ui/button";
	import {
		ChevronsUpDownIcon,
		FileIcon,
		ImageIcon,
		RotateCwSquareIcon,
		ViewIcon,
	} from "@lucide/svelte";
	import * as Tooltip from "@/components/ui/tooltip";
	import { buttonVariants } from "@/components/ui/button";
	import * as FastPng from "fast-png";
	import { toast } from "svelte-sonner";
	import * as Panzoom from "@panzoom/panzoom";
	import { resultTryCatch } from "@/result";
	import * as Dialog from "$lib/components/ui/dialog";
	import * as Drawer from "$lib/components/ui/drawer";
	import * as Collapsible from "$lib/components/ui/collapsible";
	import * as Select from "$lib/components/ui/select";
	import { MediaQuery } from "svelte/reactivity";
	import { tim3FromPng } from "@/formats/tim3";
	import { tim2FromPng } from "@/formats/tim2";

	const TIM_OPTIONS = [
		{ value: "tim3", label: "TIM3" },
		{ value: "tim2", label: "TIM2" },
	] as const;

	const BPP_OPTIONS = [
		{ value: "8bpp", label: "8 Bit Texture", disabled: false },
		{ value: "4bpp", label: "4 Bit Texture", disabled: false },
	] as const;

	let inputFile = $state<HTMLInputElement>();

	let canvas = $state<HTMLCanvasElement>();
	let panzoom = $state<Panzoom.PanzoomObject>();

	let item = $state<{ filename: string; png: FastPng.DecodedPng } | null>(null);

	const isDesktop = new MediaQuery("(min-width: 550px)");

	let isOpenConvertToTim = $state(false);
	let convertToTimFormatOptionValue = $state("tim3");
	const convertToTimFormatOptionLabel = $derived(
		TIM_OPTIONS.find((f) => f.value === convertToTimFormatOptionValue)?.label ?? "Select Format",
	);
	let convertToTimBppOptionValue = $state("8bpp");
	const convertToTimBppOptionLabel = $derived(
		BPP_OPTIONS.find((f) => f.value === convertToTimBppOptionValue)?.label ?? "Select Type",
	);

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
			const { value: png, error: pngDecodeError } = resultTryCatch(() => {
				return FastPng.decode(buffer);
			});
			if (pngDecodeError !== undefined) {
				toast.error(pngDecodeError.name, {
					description: pngDecodeError.message,
					action: {
						label: "Close",
						onClick: () => console.log(pngDecodeError),
					},
				});
				return;
			}
			if (png.palette === undefined) {
				const indexedModeError = new Error("PNG is not indexed mode");
				toast.error(indexedModeError.name, {
					description: indexedModeError.message,
					action: {
						label: "Close",
						onClick: () => console.log(indexedModeError),
					},
				});
				return;
			}

			const isTwoPixelPerByte = ~~((png.width * png.height) / 2) === png.data.length;
			const data: number[] = [];
			for (const index of png.data) {
				if (isTwoPixelPerByte) {
					const low = index & 0xf;
					const high = (index >> 4) & 0xf;
					{
						let [r, g, b, a] = png.palette[low];
						a ??= 255;
						data.push(r, g, b, a);
					}
					{
						let [r, g, b, a] = png.palette[high];
						a ??= 255;
						data.push(r, g, b, a);
					}
				} else {
					let [r, g, b, a] = png.palette[index];
					a ??= 255;
					data.push(r, g, b, a);
				}
			}

			item = {
				filename: fs[0].name.replace(/\.[^/.]+$/, ""),
				png: png,
			};

			const image = new ImageData(new Uint8ClampedArray(data), png.width, png.height);

			canvas!.width = image.width;
			canvas!.height = image.height;

			const ctx = canvas!.getContext("2d")!;
			ctx.clearRect(0, 0, image.width, image.height);
			ctx.putImageData(image, 0, 0);

			const isCan4Bpp = png.palette.length <= 16;
			// @ts-ignore
			BPP_OPTIONS[1].disabled = !isCan4Bpp;
			convertToTimBppOptionValue = isCan4Bpp ? "4bpp" : "8bpp";
		});
		r.readAsArrayBuffer(fs[0]);
	};

	const convert = () => {
		if (item === null) {
			return;
		}

		switch (convertToTimFormatOptionValue) {
			case "tim3":
				{
					const { value: tim3FromPngValue, error: tim3FromPngError } = tim3FromPng(
						convertToTimBppOptionValue,
						item.png.width,
						item.png.height,
						new Uint8Array(item.png.data),
						item.png.palette ?? [],
					);
					if (tim3FromPngError !== undefined) {
						toast.error(tim3FromPngError.name, {
							description: tim3FromPngError.message,
							action: {
								label: "Close",
								onClick: () => console.log(tim3FromPngError),
							},
						});
						return;
					}

					const blob = new Blob([new Uint8Array(tim3FromPngValue)], {
						type: "application/octet-stream",
					});

					const link = document.createElement("a");
					link.download = `${item.filename}.tm3`;
					link.href = URL.createObjectURL(blob);

					document.body.appendChild(link);
					link.click();
					document.body.removeChild(link);
				}
				break;
			case "tim2":
				{
					const { value: tim2FromPngValue, error: tim2FromPngError } = tim2FromPng(
						convertToTimBppOptionValue,
						item.png.width,
						item.png.height,
						new Uint8Array(item.png.data),
						item.png.palette ?? [],
					);
					if (tim2FromPngError !== undefined) {
						toast.error(tim2FromPngError.name, {
							description: tim2FromPngError.message,
							action: {
								label: "Close",
								onClick: () => console.log(tim2FromPngError),
							},
						});
						return;
					}

					const blob = new Blob([new Uint8Array(tim2FromPngValue)], {
						type: "application/octet-stream",
					});

					const link = document.createElement("a");
					link.download = `${item.filename}.tm2`;
					link.href = URL.createObjectURL(blob);

					document.body.appendChild(link);
					link.click();
					document.body.removeChild(link);
				}
				break;
		}
	};
</script>

<svelte:head>
	<title>TIM From PNG</title>
</svelte:head>

<!-- canvas -->
<div class="absolute inset-0 flex flex-col items-start justify-start">
	<div
		class="m-auto"
		class:opacity-0={item === null}
		style="background: repeating-conic-gradient(#cccccc 0% 25%, #f0f0f0 0% 50%) 50% / 24px 24px"
	>
		<canvas bind:this={canvas}></canvas>
	</div>
</div>

<DropZoneOverlay {ondrop} hide={item !== null} />

<!-- top left -->
<div class="absolute left-4 top-4">
	<div class="flex items-center gap-2">
		<Sidebar.Trigger />
		<div class="text-sm font-semibold tracking-tight">TIM From PNG</div>
	</div>
</div>

<!-- bottom -->
<div class="absolute bottom-0 left-0 right-0 flex flex-col">
	<div class="flex flex-row gap-2 px-4 pb-4">
		<input
			bind:this={inputFile}
			type="file"
			accept="image/png"
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
		<Button
			class="ml-auto"
			disabled={item === null}
			onclick={() => {
				isOpenConvertToTim = true;
			}}
		>
			<ImageIcon class="mr-2 size-4" />
			Convert to TIM
		</Button>
	</div>
</div>
{#if item !== null}
	{#if isDesktop.current}
		<Dialog.Root bind:open={isOpenConvertToTim}>
			<Dialog.Content class="sm:max-w-[480px]">
				<div class="flex flex-col gap-4">
					<div class="flex flex-col gap-1">
						<Collapsible.Root open={true} class="flex flex-col gap-1">
							<div class="flex flex-row gap-1">
								<div class="flex flex-col gap-1">
									<div class="text-sm font-semibold tracking-tight">CLUT Colors</div>
									<div class="text-xs font-medium text-muted-foreground">
										{item?.png?.palette?.length}
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
									{#each item?.png?.palette ?? [] as clut}
										<div
											style={`aspect-ratio: 1/1; background: #${clut[0].toString(16).padStart(2, "0")}${clut[1].toString(16).padStart(2, "0")}${clut[2].toString(16).padStart(2, "0")}${(clut[3] ?? 255).toString(16).padStart(2, "0")}`}
										></div>
									{/each}
								</div>
							</Collapsible.Content>
						</Collapsible.Root>
					</div>
					<div class="grid grid-cols-2 gap-1">
						<div class="flex flex-col gap-2">
							<div class="text-sm font-semibold tracking-tight">Format</div>
							<div>
								<Select.Root type="single" name="format" bind:value={convertToTimFormatOptionValue}>
									<Select.Trigger class="w-full">
										{convertToTimFormatOptionLabel}
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each TIM_OPTIONS as tim (tim.value)}
												<Select.Item value={tim.value} label={tim.label}>
													{tim.label}
												</Select.Item>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</div>
						</div>
						<div class="flex flex-col gap-2">
							<div class="text-sm font-semibold tracking-tight">Type</div>
							<div>
								<Select.Root type="single" name="bpp" bind:value={convertToTimBppOptionValue}>
									<Select.Trigger class="w-full">
										{convertToTimBppOptionLabel}
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each BPP_OPTIONS as bpp (bpp.value)}
												<Select.Item value={bpp.value} label={bpp.label} disabled={bpp.disabled}>
													{bpp.label}
												</Select.Item>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</div>
						</div>
					</div>
				</div>
				<Dialog.Footer>
					<Button onclick={convert}><RotateCwSquareIcon class="mr-2 size-4" /> Convert</Button>
				</Dialog.Footer>
			</Dialog.Content>
		</Dialog.Root>
	{:else}
		<Drawer.Root bind:open={isOpenConvertToTim}>
			<Drawer.Content>
				<div class="flex flex-col gap-4 px-4 pt-2">
					<div class="flex flex-col gap-1">
						<Collapsible.Root open={true} class="flex flex-col gap-1">
							<div class="flex flex-row gap-1">
								<div class="flex flex-col gap-1">
									<div class="text-sm font-semibold tracking-tight">CLUT Colors</div>
									<div class="text-xs font-medium text-muted-foreground">
										{item?.png?.palette?.length}
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
									{#each item?.png?.palette ?? [] as clut}
										<div
											style={`aspect-ratio: 1/1; background: #${clut[0].toString(16).padStart(2, "0")}${clut[1].toString(16).padStart(2, "0")}${clut[2].toString(16).padStart(2, "0")}${(clut[3] ?? 255).toString(16).padStart(2, "0")}`}
										></div>
									{/each}
								</div>
							</Collapsible.Content>
						</Collapsible.Root>
					</div>
					<div class="grid grid-cols-2 gap-1">
						<div class="flex flex-col gap-2">
							<div class="text-sm font-semibold tracking-tight">Format</div>
							<div>
								<Select.Root type="single" name="format" bind:value={convertToTimFormatOptionValue}>
									<Select.Trigger class="w-full">
										{convertToTimFormatOptionLabel}
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each TIM_OPTIONS as tim (tim.value)}
												<Select.Item value={tim.value} label={tim.label}>
													{tim.label}
												</Select.Item>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</div>
						</div>
						<div class="flex flex-col gap-2">
							<div class="text-sm font-semibold tracking-tight">Type</div>
							<div>
								<Select.Root type="single" name="bpp" bind:value={convertToTimBppOptionValue}>
									<Select.Trigger class="w-full">
										{convertToTimBppOptionLabel}
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each BPP_OPTIONS as bpp (bpp.value)}
												<Select.Item value={bpp.value} label={bpp.label} disabled={bpp.disabled}>
													{bpp.label}
												</Select.Item>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</div>
						</div>
					</div>
				</div>
				<Drawer.Footer>
					<Button onclick={convert}><RotateCwSquareIcon class="mr-2 size-4" /> Convert</Button>
					<Drawer.Close class={buttonVariants({ variant: "outline" })}>Close</Drawer.Close>
				</Drawer.Footer>
			</Drawer.Content>
		</Drawer.Root>
	{/if}
{/if}
