package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/anasrar/chihuahua/pkg/dat"
	rlig "github.com/anasrar/chihuahua/pkg/raylib_imgui"
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

func drop(filePath string) error {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var m dat.Metadata
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}
	datMetadata = &m

	writeLog(fmt.Sprintf("Entry Total: %d", m.EntryTotal))
	writeLog("Ready")

	return nil
}

func gui() {
	rl.InitWindow(int32(width), int32(height), "DAT Packer")
	defer rl.CloseWindow()
	rl.SetTargetFPS(30)

	rlig.Load()
	defer rlig.Unload()
	imgui.StyleColorsDark()

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

			if err := drop(filePath); err != nil {
				metadataPath = ""
				canPack = false
				canCancel = false
			} else {
				metadataPath = filePath
				canPack = true
				canCancel = false
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

			imgui.InternalDockBuilderDockWindow("Pack", dockUp)
			imgui.InternalDockBuilderDockWindow("Log", dockDown)

			imgui.InternalDockBuilderFinish(dock)
		}

		noTabBar := imgui.NewWindowClass()
		noTabBar.SetDockNodeFlagsOverrideSet(imgui.DockNodeFlags(imgui.DockNodeFlagsNoTabBar))

		{
			imgui.SetNextWindowClass(noTabBar)
			imgui.BeginV("Pack", nil, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoTitleBar)
			{
				imgui.BeginDisabledV(!canPack)
				if imgui.Button("Pack") {
					ctx, cancel = context.WithCancel(context.Background())

					go func() {
						if err := pack(
							ctx,
							metadataPath,
							func(total, current uint32, name string) {
								writeLog(fmt.Sprintf("%d/%d(%s): start", current, total, name))
							},
							func(total, current uint32, name string) {
								writeLog(fmt.Sprintf("%d/%d(%s): done", current, total, name))

								progress = float32(current) / float32(total)
								if total == current {
									writeLog("Done")

									progress = 0
									canPack = true
									canCancel = false
								}
							},
						); err != nil {
							writeLog(err.Error())

							progress = 0
							canPack = true
							canCancel = false
						}
					}()

					progress = 0
					canPack = false
					canCancel = true
				}
				imgui.EndDisabled()

				imgui.SameLineV(0, 4)
				imgui.BeginDisabledV(!canCancel)
				if imgui.Button("Cancel") {
					cancel()
					writeLog("Plase Wait For Cancellation")
					canPack = false
					canCancel = false
				}
				imgui.EndDisabled()

				if datMetadata != nil {
					imgui.SameLineV(0, 12)
					imgui.Text(fmt.Sprintf("Entries: %d", datMetadata.EntryTotal))

					if !canPack && canCancel {
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

				imgui.TableSetupColumnV("Type", imgui.TableColumnFlagsWidthFixed, 0, 0)
				imgui.TableSetupColumnV("Source", imgui.TableColumnFlagsWidthStretch, 2, 0)
				imgui.TableSetupColumnV("Null", imgui.TableColumnFlagsWidthStretch, 0.5, 0)
				imgui.TableSetupScrollFreeze(0, 1)
				imgui.TableHeadersRow()

				if datMetadata != nil {
					for _, entry := range datMetadata.Entries {
						imgui.TableNextRowV(imgui.TableRowFlagsNone, 0)

						imgui.TableSetColumnIndex(0)
						imgui.Text(entry.Type)

						imgui.TableSetColumnIndex(1)
						imgui.Text(entry.Source)

						imgui.TableSetColumnIndex(2)
						if entry.IsNull {
							imgui.Text("true")
						} else {
							imgui.Text("false")
						}
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
