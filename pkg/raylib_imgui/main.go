package rlig

import (
	"fmt"
	"image"
	"runtime"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

/*
int SzFloat() {
return sizeof(float);
}
*/
import "C"

type cImDrawVertx32 struct {
	Pos struct{ X, Y float32 }
	UV  struct{ X, Y float32 }
	Col uint32
}

type cVec2x64 struct {
	X float64
	Y float64
}

type cImDrawVertx64 struct {
	Pos cVec2x64
	UV  cVec2x64
	Col uint32
}

type vertex struct {
	Uv    [2]float32
	Pos   [2]float32
	Color [4]uint8
}

type clipboard struct{}

func (self clipboard) GetClipboard() string {
	return rl.GetClipboardText()
}

func (self clipboard) SetClipboard(value string) {
	rl.SetClipboardText(value)
}

type keymap struct {
	Rl int32
	Ig imgui.Key
}

var keymaps = []keymap{
	{Rl: rl.KeyApostrophe, Ig: imgui.KeyApostrophe},
	{Rl: rl.KeyComma, Ig: imgui.KeyComma},
	{Rl: rl.KeyMinus, Ig: imgui.KeyMinus},
	{Rl: rl.KeyPeriod, Ig: imgui.KeyPeriod},
	{Rl: rl.KeySlash, Ig: imgui.KeySlash},
	{Rl: rl.KeyZero, Ig: imgui.Key0},
	{Rl: rl.KeyOne, Ig: imgui.Key1},
	{Rl: rl.KeyTwo, Ig: imgui.Key2},
	{Rl: rl.KeyThree, Ig: imgui.Key3},
	{Rl: rl.KeyFour, Ig: imgui.Key4},
	{Rl: rl.KeyFive, Ig: imgui.Key5},
	{Rl: rl.KeySix, Ig: imgui.Key6},
	{Rl: rl.KeySeven, Ig: imgui.Key7},
	{Rl: rl.KeyEight, Ig: imgui.Key8},
	{Rl: rl.KeyNine, Ig: imgui.Key9},
	{Rl: rl.KeySemicolon, Ig: imgui.KeySemicolon},
	{Rl: rl.KeyEqual, Ig: imgui.KeyEqual},
	{Rl: rl.KeyA, Ig: imgui.KeyA},
	{Rl: rl.KeyB, Ig: imgui.KeyB},
	{Rl: rl.KeyC, Ig: imgui.KeyC},
	{Rl: rl.KeyD, Ig: imgui.KeyD},
	{Rl: rl.KeyE, Ig: imgui.KeyE},
	{Rl: rl.KeyF, Ig: imgui.KeyF},
	{Rl: rl.KeyG, Ig: imgui.KeyG},
	{Rl: rl.KeyH, Ig: imgui.KeyH},
	{Rl: rl.KeyI, Ig: imgui.KeyI},
	{Rl: rl.KeyJ, Ig: imgui.KeyJ},
	{Rl: rl.KeyK, Ig: imgui.KeyK},
	{Rl: rl.KeyL, Ig: imgui.KeyL},
	{Rl: rl.KeyM, Ig: imgui.KeyM},
	{Rl: rl.KeyN, Ig: imgui.KeyN},
	{Rl: rl.KeyO, Ig: imgui.KeyO},
	{Rl: rl.KeyP, Ig: imgui.KeyP},
	{Rl: rl.KeyQ, Ig: imgui.KeyQ},
	{Rl: rl.KeyR, Ig: imgui.KeyR},
	{Rl: rl.KeyS, Ig: imgui.KeyS},
	{Rl: rl.KeyT, Ig: imgui.KeyT},
	{Rl: rl.KeyU, Ig: imgui.KeyU},
	{Rl: rl.KeyV, Ig: imgui.KeyV},
	{Rl: rl.KeyW, Ig: imgui.KeyW},
	{Rl: rl.KeyX, Ig: imgui.KeyX},
	{Rl: rl.KeyY, Ig: imgui.KeyY},
	{Rl: rl.KeyZ, Ig: imgui.KeyZ},
	{Rl: rl.KeySpace, Ig: imgui.KeySpace},
	{Rl: rl.KeyEscape, Ig: imgui.KeyEscape},
	{Rl: rl.KeyEnter, Ig: imgui.KeyEnter},
	{Rl: rl.KeyTab, Ig: imgui.KeyTab},
	{Rl: rl.KeyBackspace, Ig: imgui.KeyBackspace},
	{Rl: rl.KeyInsert, Ig: imgui.KeyInsert},
	{Rl: rl.KeyDelete, Ig: imgui.KeyDelete},
	{Rl: rl.KeyRight, Ig: imgui.KeyRightArrow},
	{Rl: rl.KeyLeft, Ig: imgui.KeyLeftArrow},
	{Rl: rl.KeyDown, Ig: imgui.KeyDownArrow},
	{Rl: rl.KeyUp, Ig: imgui.KeyUpArrow},
	{Rl: rl.KeyPageUp, Ig: imgui.KeyPageUp},
	{Rl: rl.KeyPageDown, Ig: imgui.KeyPageDown},
	{Rl: rl.KeyHome, Ig: imgui.KeyHome},
	{Rl: rl.KeyEnd, Ig: imgui.KeyEnd},
	{Rl: rl.KeyCapsLock, Ig: imgui.KeyCapsLock},
	{Rl: rl.KeyScrollLock, Ig: imgui.KeyScrollLock},
	{Rl: rl.KeyNumLock, Ig: imgui.KeyNumLock},
	{Rl: rl.KeyPrintScreen, Ig: imgui.KeyPrintScreen},
	{Rl: rl.KeyPause, Ig: imgui.KeyPause},
	{Rl: rl.KeyF1, Ig: imgui.KeyF1},
	{Rl: rl.KeyF2, Ig: imgui.KeyF2},
	{Rl: rl.KeyF3, Ig: imgui.KeyF3},
	{Rl: rl.KeyF4, Ig: imgui.KeyF4},
	{Rl: rl.KeyF5, Ig: imgui.KeyF5},
	{Rl: rl.KeyF6, Ig: imgui.KeyF6},
	{Rl: rl.KeyF7, Ig: imgui.KeyF7},
	{Rl: rl.KeyF8, Ig: imgui.KeyF8},
	{Rl: rl.KeyF9, Ig: imgui.KeyF9},
	{Rl: rl.KeyF10, Ig: imgui.KeyF10},
	{Rl: rl.KeyF11, Ig: imgui.KeyF11},
	{Rl: rl.KeyF12, Ig: imgui.KeyF12},
	{Rl: rl.KeyLeftShift, Ig: imgui.KeyLeftShift},
	{Rl: rl.KeyLeftControl, Ig: imgui.KeyLeftCtrl},
	{Rl: rl.KeyLeftAlt, Ig: imgui.KeyLeftAlt},
	{Rl: rl.KeyLeftSuper, Ig: imgui.KeyLeftSuper},
	{Rl: rl.KeyRightShift, Ig: imgui.KeyRightShift},
	{Rl: rl.KeyRightControl, Ig: imgui.KeyRightCtrl},
	{Rl: rl.KeyRightAlt, Ig: imgui.KeyRightAlt},
	{Rl: rl.KeyRightSuper, Ig: imgui.KeyRightSuper},
	{Rl: rl.KeyKbMenu, Ig: imgui.KeyMenu},
	{Rl: rl.KeyLeftBracket, Ig: imgui.KeyLeftBracket},
	{Rl: rl.KeyBackSlash, Ig: imgui.KeyBackslash},
	{Rl: rl.KeyRightBracket, Ig: imgui.KeyRightBracket},
	{Rl: rl.KeyGrave, Ig: imgui.KeyGraveAccent},
	{Rl: rl.KeyKp0, Ig: imgui.KeyKeypad0},
	{Rl: rl.KeyKp1, Ig: imgui.KeyKeypad1},
	{Rl: rl.KeyKp2, Ig: imgui.KeyKeypad2},
	{Rl: rl.KeyKp3, Ig: imgui.KeyKeypad3},
	{Rl: rl.KeyKp4, Ig: imgui.KeyKeypad4},
	{Rl: rl.KeyKp5, Ig: imgui.KeyKeypad5},
	{Rl: rl.KeyKp6, Ig: imgui.KeyKeypad6},
	{Rl: rl.KeyKp7, Ig: imgui.KeyKeypad7},
	{Rl: rl.KeyKp8, Ig: imgui.KeyKeypad8},
	{Rl: rl.KeyKp9, Ig: imgui.KeyKeypad9},
	{Rl: rl.KeyKpDecimal, Ig: imgui.KeyKeypadDecimal},
	{Rl: rl.KeyKpDivide, Ig: imgui.KeyKeypadDivide},
	{Rl: rl.KeyKpMultiply, Ig: imgui.KeyKeypadMultiply},
	{Rl: rl.KeyKpSubtract, Ig: imgui.KeyKeypadSubtract},
	{Rl: rl.KeyKpAdd, Ig: imgui.KeyKeypadAdd},
	{Rl: rl.KeyKpEnter, Ig: imgui.KeyKeypadEnter},
	{Rl: rl.KeyKpEqual, Ig: imgui.KeyKeypadEqual},
}

var sizeoffloat int
var GlobalContext *imgui.Context = imgui.CreateContext()
var FontTexture *rl.Texture2D

var LastFrameFocused = false

var CurrentMouseCursor imgui.MouseCursor = imgui.MouseCursorCOUNT

// NOTE: mouse state last frame
var LastMouseLeftPressed = false
var LastMouseRightPressed = false
var LastMouseMiddlePressed = false
var LastMouseForwardPressed = false
var LastMouseBackPressed = false

// NOTE: keyboard modifier last frame
var LastKeyControlPressed = false
var LastKeyShiftPressed = false
var LastKeyAltPressed = false
var LastKeySuperPressed = false

func init() {
	sizeoffloat = int(C.SzFloat())
}

func getVertices(vbuf unsafe.Pointer, vblen, vsize, offpos, offuv,
	offcol int,
) []vertex {
	if sizeoffloat == 4 {
		return getVerticesx32(vbuf, vblen, vsize, offpos, offuv, offcol)
	}
	if sizeoffloat == 8 {
		return getVerticesx64(vbuf, vblen, vsize, offpos, offuv, offcol)
	}
	panic("invalid char size")
}

func getVerticesx32(vbuf unsafe.Pointer, vblen, vsize, offpos, offuv,
	offcol int,
) []vertex {
	n := vblen / vsize
	vertices := make([]vertex, 0, vblen/vsize)
	if offpos != 0 || offuv != 8 || offcol != 16 {
		panic("TODO: invalid vertex layout")
	}

	rawverts := (*[1 << 28]cImDrawVertx32)(vbuf)[:n:n]
	for i := 0; i < n; i++ {
		vertices = append(vertices, vertex{
			Uv: [2]float32{
				rawverts[i].UV.X,
				rawverts[i].UV.Y,
			},
			Pos: [2]float32{
				rawverts[i].Pos.X,
				rawverts[i].Pos.Y,
			},
			Color: [4]uint8{
				uint8(rawverts[i].Col & 0xFF),
				uint8(rawverts[i].Col >> 8 & 0xFF),
				uint8(rawverts[i].Col >> 16 & 0xFF),
				uint8(rawverts[i].Col >> 24 & 0xFF),
			},
		})
	}
	return vertices
}

func getVerticesx64(vbuf unsafe.Pointer, vblen, vsize, offpos, offuv,
	offcol int,
) []vertex {
	n := vblen / vsize
	vertices := make([]vertex, 0, vblen/vsize)
	if offpos != 0 || offuv != 8 || offcol != 16 {
		panic("TODO: invalid vertex layout (64)")
	}
	rawverts := (*[1 << 28]cImDrawVertx64)(vbuf)[:n:n]
	for i := 0; i < n; i++ {
		vertices = append(vertices, vertex{
			Uv: [2]float32{
				float32(rawverts[i].UV.X),
				float32(rawverts[i].UV.Y),
			},
			Pos: [2]float32{
				float32(rawverts[i].Pos.X),
				float32(rawverts[i].Pos.Y),
			},
			Color: [4]uint8{
				uint8(rawverts[i].Col & 0xFF),
				uint8(rawverts[i].Col >> 8 & 0xFF),
				uint8(rawverts[i].Col >> 16 & 0xFF),
				uint8(rawverts[i].Col >> 24 & 0xFF),
			},
		})
	}
	return vertices
}

func getIndices(ibuf unsafe.Pointer, iblen, isize int) []uint16 {
	n := iblen / isize
	switch isize {
	case 2:
		// direct conversion (without a data copy)
		// TODO: document the size limit (?) this fits 268435456 bytes
		// https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
		return (*[1 << 28]uint16)(ibuf)[:n:n]
	case 4:
		slc := make([]uint16, n)
		for i := 0; i < n; i++ {
			slc[i] = uint16(*(*uint32)(unsafe.Pointer(uintptr(ibuf) + uintptr(i*isize))))
		}
		return slc
	case 8:
		slc := make([]uint16, n)
		for i := 0; i < n; i++ {
			slc[i] = uint16(*(*uint64)(unsafe.Pointer(uintptr(ibuf) + uintptr(i*isize))))
		}
		return slc
	default:
		panic(fmt.Sprint("byte size", isize, "not supported"))
	}
}

// TODO: upgrade raylib-go and change rlButton int32 to rlButton rl.MouseButton
func mouseButtonUpdate(lastFrame *bool, rlButton int32, igButton int32) {
	val := rl.IsMouseButtonDown(rlButton)
	if val != *lastFrame {
		imgui.CurrentIO().AddMouseButtonEvent(igButton, val)
	}
	*lastFrame = val
}

func keyModifierButtonUpdate(lastFrame *bool, rlButtonLeft, rlButtonRight int32, igButton imgui.Key) {
	val := rl.IsKeyDown(rlButtonLeft) || rl.IsKeyDown(rlButtonRight)
	if val != *lastFrame {
		imgui.CurrentIO().AddKeyEvent(igButton, val)
	}
	*lastFrame = val
}

func gamePadButtonUpdate(rlButton int32, igButton imgui.Key) {
	if rl.IsGamepadButtonPressed(0, rlButton) {
		imgui.CurrentIO().AddKeyEvent(igButton, true)
	} else if rl.IsGamepadButtonReleased(0, rlButton) {
		imgui.CurrentIO().AddKeyEvent(igButton, false)
	}
}

func gamePadAnalogUpdate(axis int32, igAnalogNeg, igAnalogPos imgui.Key) {
	const deadZone = 0.20
	value := rl.GetGamepadAxisMovement(0, axis)

	negValue := float32(0)
	if value < -deadZone {
		negValue = -value
	}
	imgui.CurrentIO().AddKeyAnalogEvent(igAnalogNeg, value < -deadZone, negValue)

	posValue := float32(0)
	if value > deadZone {
		posValue = value
	}
	imgui.CurrentIO().AddKeyAnalogEvent(igAnalogPos, value > deadZone, posValue)
}

func Load() {
	imgui.SetCurrentContext(GlobalContext)
	imgui.CurrentIO().SetConfigFlags(imgui.ConfigFlagsNavEnableKeyboard)
	imgui.CurrentIO().SetBackendPlatformName("cimgui_go_impl_raylib_go")
	imgui.CurrentIO().SetBackendFlags(
		imgui.BackendFlagsHasMouseCursors | imgui.BackendFlagsHasSetMousePos,
	)
	imgui.CurrentIO().SetConfigFlags(
		imgui.ConfigFlagsNavEnableKeyboard | imgui.ConfigFlagsDockingEnable,
	)
	imgui.CurrentIO().SetIniFilename("")

	imgui.CurrentPlatformIO().SetClipboardHandler(clipboard{})

	// NOTE: default font atlas
	{
		fonts := imgui.CurrentIO().Fonts()
		pixels, width, height, _ := fonts.GetTextureDataAsRGBA32()
		img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		for y := int32(0); y < height; y++ {
			for x := int32(0); x < width; x++ {
				offset := (y*width + x) * 4
				ptr := (*[4]byte)(unsafe.Pointer(uintptr(unsafe.Pointer(pixels)) + uintptr(offset)))
				imgOffset := (y*width + x) * 4
				img.Pix[imgOffset+0] = ptr[0]
				img.Pix[imgOffset+1] = ptr[1]
				img.Pix[imgOffset+2] = ptr[2]
				img.Pix[imgOffset+3] = ptr[3]
			}
		}
		rlImage := rl.NewImageFromImage(img)
		defer rl.UnloadImage(rlImage)
		FontTexture := rl.LoadTextureFromImage(rlImage)
		fonts.SetTexID(imgui.TextureID(FontTexture.ID))
	}
}

func Unload() {
	imgui.DestroyContextV(GlobalContext)
	if FontTexture != nil {
		rl.UnloadTexture(*FontTexture)
	}
}

func Update() {
	focused := rl.IsWindowFocused()
	if focused != LastFrameFocused {
		imgui.CurrentIO().AddFocusEvent(focused)
	}
	LastFrameFocused = focused

	// NOTE: window size bind
	if rl.IsWindowFullscreen() {
		monitor := rl.GetCurrentMonitor()
		imgui.CurrentIO().SetDisplaySize(imgui.NewVec2(float32(rl.GetMonitorWidth(monitor)), float32(rl.GetMonitorHeight(monitor))))
	} else {
		imgui.CurrentIO().SetDisplaySize(imgui.NewVec2(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight())))
	}

	// NOTE: dpi scale bind
	resolutionScale := rl.GetWindowScaleDPI()
	if runtime.GOOS == "darwin" {
		resolutionScale = rl.Vector2One()
	}
	imgui.CurrentIO().SetDisplayFramebufferScale(imgui.NewVec2(resolutionScale.X, resolutionScale.Y))

	// NOTE: set delta time
	imgui.CurrentIO().SetDeltaTime(rl.GetFrameTime())

	// NOTE: cursor icon
	if (imgui.CurrentIO().BackendFlags() & imgui.BackendFlagsHasMouseCursors) != 0 {
		if imgui.CurrentIO().ConfigFlags()&imgui.ConfigFlagsNoMouseCursorChange == 0 {
			cursor := imgui.CurrentMouseCursor()
			if cursor != CurrentMouseCursor || imgui.CurrentIO().MouseDrawCursor() {
				CurrentMouseCursor = cursor
				if imgui.CurrentIO().MouseDrawCursor() || cursor == imgui.MouseCursorNone {
					rl.HideCursor()
				} else {
					rl.ShowCursor()

					if imgui.CurrentIO().ConfigFlags()&imgui.ConfigFlagsNoMouseCursorChange != 1 {
						value := rl.MouseCursorDefault
						if cursor > -1 && cursor < imgui.MouseCursorCOUNT {
							switch cursor {
							case imgui.MouseCursorArrow:
								value = rl.MouseCursorArrow
							case imgui.MouseCursorTextInput:
								value = rl.MouseCursorIBeam
							case imgui.MouseCursorHand:
								value = rl.MouseCursorPointingHand
							case imgui.MouseCursorResizeAll:
								value = rl.MouseCursorResizeAll
							case imgui.MouseCursorResizeEW:
								value = rl.MouseCursorResizeEW
							case imgui.MouseCursorResizeNESW:
								value = rl.MouseCursorResizeNESW
							case imgui.MouseCursorResizeNS:
								value = rl.MouseCursorResizeNS
							case imgui.MouseCursorResizeNWSE:
								value = rl.MouseCursorResizeNWSE
							case imgui.MouseCursorNotAllowed:
								value = rl.MouseCursorNotAllowed
							}
						}

						rl.SetMouseCursor(value)
					}
				}
			}
		}
	}

	// NOTE: mouse position bind
	if imgui.CurrentIO().WantSetMousePos() {
		mousePosition := imgui.CurrentIO().MousePos()
		rl.SetMousePosition(int(mousePosition.X), int(mousePosition.Y))
	} else {
		mousePosition := rl.GetMousePosition()
		imgui.CurrentIO().SetMousePos(imgui.NewVec2(mousePosition.X, mousePosition.Y))
	}

	// NOTE: mouse button bind
	mouseButtonUpdate(&LastMouseLeftPressed, rl.MouseButtonLeft, int32(imgui.MouseButtonLeft))
	mouseButtonUpdate(&LastMouseRightPressed, rl.MouseButtonRight, int32(imgui.MouseButtonRight))
	mouseButtonUpdate(&LastMouseMiddlePressed, rl.MouseButtonMiddle, int32(imgui.MouseButtonMiddle))
	mouseButtonUpdate(&LastMouseForwardPressed, rl.MouseButtonForward, int32(imgui.MouseButtonMiddle)+1)
	mouseButtonUpdate(&LastMouseBackPressed, rl.MouseButtonBack, int32(imgui.MouseButtonMiddle)+2)

	// NOTE: mouse wheel bind
	mouseWheel := rl.GetMouseWheelMoveV()
	imgui.CurrentIO().AddMouseWheelEvent(mouseWheel.X, mouseWheel.Y)

	// NOTE: keyboard modifier
	keyModifierButtonUpdate(&LastKeyControlPressed, rl.KeyRightControl, rl.KeyLeftControl, imgui.ModCtrl)
	keyModifierButtonUpdate(&LastKeyShiftPressed, rl.KeyRightShift, rl.KeyLeftShift, imgui.ModShift)
	keyModifierButtonUpdate(&LastKeyAltPressed, rl.KeyRightAlt, rl.KeyLeftAlt, imgui.ModAlt)
	keyModifierButtonUpdate(&LastKeySuperPressed, rl.KeyRightSuper, rl.KeyLeftSuper, imgui.ModSuper)

	// NOTE: keyboard button
	for _, entry := range keymaps {
		if rl.IsKeyReleased(entry.Rl) {
			imgui.CurrentIO().AddKeyEvent(entry.Ig, false)
		} else if rl.IsKeyPressed(entry.Rl) {
			imgui.CurrentIO().AddKeyEvent(entry.Ig, true)
		}
	}

	// NOTE: imgui input text
	if imgui.CurrentIO().WantCaptureKeyboard() {
		pressed := rl.GetCharPressed()
		for pressed != 0 {
			imgui.CurrentIO().AddInputCharacter(uint32(pressed))
			pressed = rl.GetCharPressed()
		}
	}

	// NOTE: game pad
	if imgui.CurrentIO().ConfigFlags()&imgui.ConfigFlagsNavEnableGamepad == 1 && rl.IsGamepadAvailable(0) {
		gamePadButtonUpdate(rl.GamepadButtonLeftFaceUp, imgui.KeyGamepadDpadUp)
		gamePadButtonUpdate(rl.GamepadButtonLeftFaceRight, imgui.KeyGamepadDpadRight)
		gamePadButtonUpdate(rl.GamepadButtonLeftFaceDown, imgui.KeyGamepadDpadDown)
		gamePadButtonUpdate(rl.GamepadButtonLeftFaceLeft, imgui.KeyGamepadDpadLeft)

		gamePadButtonUpdate(rl.GamepadButtonRightFaceUp, imgui.KeyGamepadFaceUp)
		gamePadButtonUpdate(rl.GamepadButtonRightFaceRight, imgui.KeyGamepadFaceLeft)
		gamePadButtonUpdate(rl.GamepadButtonRightFaceDown, imgui.KeyGamepadFaceDown)
		gamePadButtonUpdate(rl.GamepadButtonRightFaceLeft, imgui.KeyGamepadFaceRight)

		gamePadButtonUpdate(rl.GamepadButtonLeftTrigger1, imgui.KeyGamepadL1)
		gamePadButtonUpdate(rl.GamepadButtonLeftTrigger2, imgui.KeyGamepadL2)
		gamePadButtonUpdate(rl.GamepadButtonRightTrigger1, imgui.KeyGamepadR1)
		gamePadButtonUpdate(rl.GamepadButtonRightTrigger2, imgui.KeyGamepadR2)
		gamePadButtonUpdate(rl.GamepadButtonLeftThumb, imgui.KeyGamepadL3)
		gamePadButtonUpdate(rl.GamepadButtonRightThumb, imgui.KeyGamepadR3)

		gamePadButtonUpdate(rl.GamepadButtonMiddleLeft, imgui.KeyGamepadStart)
		gamePadButtonUpdate(rl.GamepadButtonMiddleRight, imgui.KeyGamepadBack)

		gamePadAnalogUpdate(rl.GamepadAxisLeftX, imgui.KeyGamepadLStickLeft, imgui.KeyGamepadLStickRight)
		gamePadAnalogUpdate(rl.GamepadAxisLeftY, imgui.KeyGamepadLStickUp, imgui.KeyGamepadLStickDown)

		gamePadAnalogUpdate(rl.GamepadAxisRightX, imgui.KeyGamepadRStickLeft, imgui.KeyGamepadRStickRight)
		gamePadAnalogUpdate(rl.GamepadAxisRightY, imgui.KeyGamepadRStickUp, imgui.KeyGamepadRStickDown)
	}
}

func Render() {
	rl.DrawRenderBatchActive()
	rl.DisableBackfaceCulling()

	imgui.Render()
	igDrawData := imgui.CurrentDrawData()
	igDisplayPosition := igDrawData.DisplayPos()

	vertexSize,
		vertexOffsetPos,
		vertexOffsetUv,
		vertexOffsetCol := imgui.VertexBufferLayout()
	indexSize := imgui.IndexBufferLayout()

	for _, cmdlist := range igDrawData.CommandLists() {
		vertexBuffer, vertexLen := cmdlist.GetVertexBuffer()
		indexBuffer, indexLen := cmdlist.GetIndexBuffer()
		vertices := getVertices(vertexBuffer, vertexLen, vertexSize, vertexOffsetPos,
			vertexOffsetUv, vertexOffsetCol)
		indices := getIndices(indexBuffer, indexLen, indexSize)
		for _, cmd := range cmdlist.Commands() {
			if cmd.HasUserCallback() {
				cmd.CallUserCallback(cmdlist)
			} else {
				clipRect := cmd.ClipRect()
				x := clipRect.X - igDisplayPosition.X
				y := clipRect.Y - igDisplayPosition.Y
				width := clipRect.Z - x
				height := clipRect.W - y

				rl.EnableScissorTest()
				scale := imgui.CurrentIO().DisplayFramebufferScale()
				rl.Scissor(int32(x*scale.X), int32(imgui.CurrentIO().DisplaySize().Y-(y+height)*scale.Y), int32(width*scale.X), int32(height*scale.Y))

				count := cmd.ElemCount()
				indexStart := cmd.IdxOffset()
				textureId := cmd.TextureId()

				if !(count < 3) {
					rl.Begin(rl.Triangles)
					rl.SetTexture(uint32(textureId))

					for i := uint32(0); i <= (count - 3); i += 3 {
						indexA := indices[indexStart+i]
						indexB := indices[indexStart+i+1]
						indexC := indices[indexStart+i+2]

						vertexA := vertices[indexA]
						vertexB := vertices[indexB]
						vertexC := vertices[indexC]

						rl.Color4ub(vertexA.Color[0], vertexA.Color[1], vertexA.Color[2], vertexA.Color[3])
						rl.TexCoord2f(vertexA.Uv[0], vertexA.Uv[1])
						rl.Vertex2f(vertexA.Pos[0], vertexA.Pos[1])

						rl.Color4ub(vertexB.Color[0], vertexB.Color[1], vertexB.Color[2], vertexB.Color[3])
						rl.TexCoord2f(vertexB.Uv[0], vertexB.Uv[1])
						rl.Vertex2f(vertexB.Pos[0], vertexB.Pos[1])

						rl.Color4ub(vertexC.Color[0], vertexC.Color[1], vertexC.Color[2], vertexC.Color[3])
						rl.TexCoord2f(vertexC.Uv[0], vertexC.Uv[1])
						rl.Vertex2f(vertexC.Pos[0], vertexC.Pos[1])
					}

					rl.End()
					rl.DrawRenderBatchActive()
				}
			}
		}
	}

	rl.SetTexture(0)
	rl.DisableScissorTest()
	rl.EnableBackfaceCulling()
}
