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

### roomviewer

Room viewer for rXXX.dat file, drag and drop `rXXX.dat` file, support export as GLTF.

### scrviewer

SCR viewer for view SCR and MD file, drag and drop SCR, MD, and TM3 file, support export as GLTF.

### t32viewer

T32 viewer for view T32 file that use as texture UI, support export as PNG.

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

- [ ] png2tim2
- [ ] png2tim3
- [ ] png2t32
- [ ] mot2gltf
- [ ] Blender Add-ons: export as SCR room
- [ ] Blender Add-ons: export as SCR model
- [ ] Blender Add-ons: import SCR room
- [ ] Blender Add-ons: import SCR model
- [ ] Blender Add-ons: import MOT model

## File Format

```
SCP: generic dat container, contain SCR and TM3.
T32: texture UI.
MDB: bones, texture index, and vertex buffer.
SCR: container for MDB, contain name and transform.
TIM2: PS2 texture format.
TIM3: extend version PS2 texture format.
TM3: TIM3 container.
MOT: contain animation curve with bone target and channel.
```

`TODO: add more file format`.

## File Name Pattern

```
orivia_X: TIM2 image texture on pause menu.
rXXX: (stage) room data.
plXX: playable character data.
idXX: generic dat contain UI stuff.
```

`TODO: add more file name pattern`.

## Built With

- https://go.dev/
- https://www.raylib.com/
- https://github.com/gen2brain/raylib-go
