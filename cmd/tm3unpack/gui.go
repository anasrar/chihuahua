package main

import (
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
	"github.com/anasrar/chihuahua/pkg/tm3"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func writeLog(msg string) {
	logs += msg
	logs += "\n"
	logUpdate = true
}

func clearLog() {
	logs = ""
}

func gui() {
	rl.InitWindow(int32(width), int32(height), "TM3 Unpacker")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()

	createDock := true

	for !rl.WindowShouldClose() {
		rlig.Update()

		if rl.IsWindowResized() {
			width = float32(rl.GetScreenWidth())
			height = float32(rl.GetScreenHeight())
		}

		if rl.IsFileDropped() {
			filePath := rl.LoadDroppedFiles()[0]
			defer rl.UnloadDroppedFiles()

			tim := tm3.New()
			if err := tm3.FromPath(tim, filePath); err != nil {
				writeLog(err.Error())

				tm3Path = ""
				canUnpack = false
				canCancel = false
				tm3Data = nil
			} else {
				writeLog(fmt.Sprintf("Entry Total: %d", tim.EntryTotal))
				writeLog("Ready")

				tm3Path = filePath
				canUnpack = true
				canCancel = false
				tm3Data = tim
			}
		}

		imgui.NewFrame()

		dock := imgui.DockSpaceOverViewport()
		if createDock {
			createDock = false
			imgui.InternalDockBuilderRemoveNode(dock)
			imgui.InternalDockBuilderAddNodeV(dock, imgui.DockNodeFlagsNone)
			imgui.InternalDockBuilderSetNodeSize(dock, imgui.MainViewport().Size())

			dockUp := imgui.InternalDockBuilderSplitNode(dock, imgui.DirUp, 0.7, nil, &dock)
			dockDown := imgui.InternalDockBuilderSplitNode(dock, imgui.DirDown, 0.3, nil, &dock)

			imgui.InternalDockBuilderDockWindow("Unpack", dockUp)
			imgui.InternalDockBuilderDockWindow("Log", dockDown)

			imgui.InternalDockBuilderFinish(dock)
		}

		noTabBar := imgui.NewWindowClass()
		noTabBar.SetDockNodeFlagsOverrideSet(imgui.DockNodeFlags(imgui.DockNodeFlagsNoTabBar))

		{
			imgui.SetNextWindowClass(noTabBar)
			imgui.BeginV("Unpack", nil, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
			{
				imgui.BeginDisabledV(!canUnpack)
				if imgui.Button("Unpack") {
					ctx, cancel = context.WithCancel(context.Background())

					go func() {
						if err := unpack(
							ctx,
							tm3Path,
							func(total, current uint32, name string) {
								writeLog(fmt.Sprintf("%d/%d(%s): start", current, total, name))
							},
							func(total, current uint32, name string) {
								writeLog(fmt.Sprintf("%d/%d(%s): done", current, total, name))

								progress = float32(current) / float32(total)
								if total == current {
									writeLog("Done")

									progress = 0
									canUnpack = true
									canCancel = false
								}
							},
						); err != nil {
							writeLog(err.Error())

							progress = 0
							canUnpack = true
							canCancel = false
						}
					}()

					progress = 0
					canUnpack = false
					canCancel = true
				}
				imgui.EndDisabled()

				imgui.SameLineV(0, 4)
				imgui.BeginDisabledV(!canCancel)
				if imgui.Button("Cancel") {
					cancel()
					writeLog("Plase Wait For Cancellation")
					canUnpack = false
					canCancel = false
				}
				imgui.EndDisabled()

				imgui.SameLineV(0, 12)
				if imgui.Button("Offset Unit") {
					switch offsetUnit {
					case OffsetUnitDecimal:
						offsetUnit = OffsetUnitHex
					case OffsetUnitHex:
						offsetUnit = OffsetUnitDecimal
					}
				}

				if tm3Data != nil {
					imgui.SameLineV(0, 12)
					imgui.Text(fmt.Sprintf("Entries: %d", tm3Data.EntryTotal))

					if !canUnpack && canCancel {
						imgui.SameLineV(0, 12)
						imgui.Text(fmt.Sprintf("Progress: %.f%%%%", progress*100))
					}
				}

			}

			{
				imgui.BeginTableV(
					"datData",
					3,
					imgui.TableFlagsRowBg|
						imgui.TableFlagsScrollY|
						imgui.TableFlagsBorders|
						imgui.TableFlagsSizingStretchSame,
					imgui.NewVec2(0, 0),
					0,
				)

				imgui.TableSetupColumnV("Name", imgui.TableColumnFlagsWidthStretch, 0.5, 0)
				imgui.TableSetupColumnV("Offset", imgui.TableColumnFlagsWidthStretch, 2, 0)
				imgui.TableSetupColumnV("Size", imgui.TableColumnFlagsWidthStretch, 1.2, 0)
				imgui.TableSetupScrollFreeze(0, 1)
				imgui.TableHeadersRow()

				if tm3Data != nil {
					for i, entry := range tm3Data.Entries {
						imgui.TableNextRowV(imgui.TableRowFlagsNone, 0)

						imgui.TableSetColumnIndex(0)
						imgui.Text(entry.Name)

						imgui.TableSetColumnIndex(1)
						imgui.PushIDInt(int32(i))
						if imgui.Button("Copy") {
							switch offsetUnit {
							case OffsetUnitDecimal:
								rl.SetClipboardText(
									fmt.Sprintf("%d", entry.Offset),
								)
							case OffsetUnitHex:
								rl.SetClipboardText(
									fmt.Sprintf("0x%X", entry.Offset),
								)
							}
						}
						imgui.PopID()

						imgui.SameLineV(0, 4)
						switch offsetUnit {
						case OffsetUnitDecimal:
							imgui.Text(fmt.Sprintf("%d", entry.Offset))
						case OffsetUnitHex:
							imgui.Text(fmt.Sprintf("0x%X", entry.Offset))
						}

						imgui.TableSetColumnIndex(2)
						imgui.Text(fmt.Sprintf("%d", entry.Size))
					}
				}

				imgui.EndTable()
			}

			imgui.End()
		}

		{
			imgui.SetNextWindowClass(noTabBar)
			imgui.BeginV("Log", nil, imgui.WindowFlagsNone|imgui.WindowFlagsNoTitleBar)
			{
				if imgui.Button("Clear") {
					clearLog()
				}
				imgui.SameLineV(0, 4)
				imgui.Checkbox("Auto Scroll", &logAutoScroll)
				imgui.BeginChildStrV("LogRegion", imgui.NewVec2(0, 0), imgui.ChildFlagsNavFlattened, imgui.WindowFlagsHorizontalScrollbar)
				imgui.Text(logs)
				if logAutoScroll && logUpdate {
					imgui.SetScrollHereYV(1)
				}
				logUpdate = false
				imgui.EndChild()
			}
			imgui.End()
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.NewColor(0x12, 0x12, 0x12, 0xFF))
		rlig.Render()
		rl.EndDrawing()

	}
}
