# Chihuahua

**Chihuahua** is project for view and modified God Hand (PS2 game) develop by Clover Studio and publish by Capcom.

## Other Projects

Since God Hand using the same game engine for Okami (the fact is some Okami asset found in God Hand USA ISO), there some overlap project that have same mission. Also after Clover Studio shut down and founded PlatinumGames, PlatinumGames still use same tech with slightly modification.

- https://github.com/Al-Hydra/GodHand-Noesis-Plugin
- https://github.com/akitotheanimator/God-Hand-Tools/
- https://github.com/christianmateus?tab=repositories&q=god+hand&type=&language=&sort=
- https://github.com/Shintensu/OkamiHD-Reverse-Engineering
- https://github.com/whataboutclyde/okami-utils
- https://github.com/allogic/Nippon
- https://github.com/Kerilk/bayonetta_tools/
- https://github.com/WoefulWolf/NieR2Blender2NieR

`TODO: add more projects`.

## Tools

### datpack

Pack generic dat container, support CLI and GUI.

### datunpack

Unpack generic dat container, support CLI and GUI.

### modelviewer

Model viewer for XXX.dat file except `evXXX.dat`, drag and drop `XXX.dat` file, support export as GLTF.

### png2tim

Convert PNG to TIM (TIM3 and TIM2), support CLI and GUI, drag and drop PNG.

> [!IMPORTANT]  
> Only PNG with indexed mode that can be converted and the colors is less than `256`.
>
> Width and height should be `2^n`, where max `n` is 9 (to not exceeded PS2 RAM).

> [!TIP]
> Convert PNG to indexed mode.
>
> GIMP: Image > Mode > Indexed.
>
> Photoshop: Image > Mode > Indexed Color.

### roomviewer

Room viewer for rXXX.dat file, drag and drop `rXXX.dat` file, support export as GLTF.

### scrviewer

SCR viewer for view SCR and MD file, drag and drop SCR, MD, and TM3 file, support export as GLTF.

### t32viewer

T32 viewer for view T32 file that use as texture UI, support export as PNG and convert PNG to T32.

> [!IMPORTANT]  
> Only PNG with indexed mode that can be converted and the colors is less than `256`.
>
> Width and height should be `2^n`, where max `n` is 9 (to not exceeded PS2 RAM).

> [!TIP]
> Convert PNG to indexed mode.
>
> GIMP: Image > Mode > Indexed.
>
> Photoshop: Image > Mode > Indexed Color.

### timviewer

TIM viewer for view TIM2 (`orivia_`), TIM3, TM3 image texture, support export as PNG.

### tm3pack

Pack TIM3 container as TM3, support CLI and GUI.

### tm3unpack

Unpack TIM3 container as TM3, support CLI and GUI.

## Developer

### ImHex

Using https://github.com/WerWolv/ImHex to analyze file format. There `pkg/*/*.hexpat` file.

## TODOS

- [ ] mot2gltf
- [ ] Blender Add-ons: export as SCR room
- [ ] Blender Add-ons: export as SCR model

## File Format

```
AFS: archive file system.
AKG: see pkg/akg/akg.hexpat.
AKT: generic dat container, contain AKG.
CMP: generic dat container, contain MOT.
EFF: generic dat container, contain TBL and EMD.
EMS: npc spawn transform.
ENV: TIM2.
FST: file system tree.
MDB: bones, texture index, and vertex buffer.
OMS: object spawn transform.
SCP: generic dat container, contain SCR and TM3.
SCR: container for MDB, contain name and transform.
MOT: contain animation curve with bone target and channel.
T32: texture UI.
TIM2: PS2 texture format.
TIM3: extend version PS2 texture format.
TM3: TIM3 container.
```

`TODO: add more file format`.

## File Name Pattern

```
elXX: model for cutscene.
emXX: npc model data.
idXX: generic dat contain UI stuff.
olXX: object model data (breakable, weapon).
orivia_X: TIM2 image texture on pause menu.
plXX: playable character data.
rXXX: (stage) room data.
```

`TODO: add more file name pattern`.

## Built With

- https://go.dev/
- https://www.raylib.com/
- https://github.com/gen2brain/raylib-go
- https://github.com/ocornut/imgui
- https://github.com/cimgui/cimgui
- https://github.com/AllenDang/cimgui-go
