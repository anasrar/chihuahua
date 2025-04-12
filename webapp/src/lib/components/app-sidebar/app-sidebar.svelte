<script lang="ts">
	import { base } from "$app/paths";
	import * as Sidebar from "$lib/components/ui/sidebar";
	import * as Collapsible from "$lib/components/ui/collapsible/index.js";
	import { ChevronRightIcon, HouseIcon, ImageIcon, ImagesIcon } from "@lucide/svelte";

	type Item = {
		title: string;
		icon: typeof ImageIcon;
		hide: boolean;
		subitems: {
			hide: boolean;
			title: string;
			url: string;
		}[];
	};

	const items: Item[] = [
		{
			title: "TM3",
			icon: ImagesIcon,
			hide: true,
			subitems: [
				{
					hide: true,
					title: "Unpack",
					url: "tm3/unpack",
				},
				{
					hide: true,
					title: "Repack",
					url: "tm3/repack",
				},
			],
		},
		{
			title: "TIM",
			icon: ImageIcon,
			hide: false,
			subitems: [
				{
					hide: false,
					title: "Viewer",
					url: "tim/viewer",
				},
				{
					hide: false,
					title: "From PNG",
					url: "tim/frompng",
				},
			],
		},
		{
			title: "T32",
			icon: ImageIcon,
			hide: true,
			subitems: [
				{
					hide: true,
					title: "Viewer",
					url: "t32/viewer",
				},
				{
					hide: true,
					title: "From PNG",
					url: "t32/frompng",
				},
			],
		},
	];
</script>

<Sidebar.Root collapsible="icon" class="font-inter">
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					<Sidebar.MenuItem>
						<Sidebar.MenuButton>
							{#snippet child({ props })}
								<a href={`${base}/`} {...props}>
									<HouseIcon />
									<span>Home</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
		<Sidebar.Separator />
		<Sidebar.Group>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					{#each items as item (item.title)}
						{#if !item.hide}
							<Collapsible.Root open class="group/collapsible">
								<Sidebar.MenuItem>
									<Collapsible.Trigger>
										{#snippet child({ props })}
											<Sidebar.MenuButton {...props}>
												{#snippet tooltipContent()}
													{item.title}
												{/snippet}
												<item.icon />
												<span>{item.title}</span>
												<ChevronRightIcon
													class="ml-auto transition-transform group-data-[state=open]/collapsible:rotate-90"
												/>
											</Sidebar.MenuButton>
										{/snippet}
									</Collapsible.Trigger>
									<Collapsible.Content>
										<Sidebar.MenuSub>
											{#each item.subitems as subitem (subitem.title)}
												{#if !subitem.hide}
													<Sidebar.MenuSubItem>
														<Sidebar.MenuButton>
															{#snippet child({ props })}
																<a href={`${base}/${subitem.url}`} {...props}>
																	<span>{subitem.title}</span>
																</a>
															{/snippet}
														</Sidebar.MenuButton>
													</Sidebar.MenuSubItem>
												{/if}
											{/each}
										</Sidebar.MenuSub>
									</Collapsible.Content>
								</Sidebar.MenuItem>
							</Collapsible.Root>
						{/if}
					{/each}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
</Sidebar.Root>
