// vi:ft=c

#define MDB_OFFSET 0x25B20 // pl00.dat
// #define MDB_OFFSET 0xCF840 // r006.dat

struct Bone {
  float x [[color("FF0000")]];
  float y [[color("00FF00")]];
  float z [[color("0000FF")]];
  padding[2];
  s16   up;
};

struct Vertex {
  float x [[color("FF0000")]];
  float y [[color("00FF00")]];
  float z [[color("0000FF")]];
  u32   flag [[color("00FFFF")]];
};

struct Normal {
  u8 x [[color("FF0000")]];
  u8 y [[color("00FF00")]];
  u8 z [[color("0000FF")]];
  padding[1];
};

struct Uv {
  u16 u [[color("FF0000")]];
  u16 v [[color("00FF00")]];
};

struct Color {
  u8 r [[color("FF0000")]];
  u8 g [[color("00FF00")]];
  u8 b [[color("0000FF")]];
  padding[1];
};

struct Weight {
  u8 indices[4] [[color("FF00FF")]];
  u8 x [[color("FF0000")]];
  u8 y [[color("00FF00")]];
  u8 z [[color("0000FF")]];
  u8 w [[color("00FFFF")]];
};

struct DataAddress {
  u32 position_offset; // NOTE: all relative to DataAddress
  u32 normal_offset;
  u32 uv_offset;
  u32 color_offset;
  u32 weight_offset;
  u16 vertex_total;
  u16 material_index;
};

struct Mdb {
  char signature[4];
  u32  bone_offset; // NOTE: relative to signature
  u16  bone_total;
  u16  data_total;
  padding[18];
  u16  flag;
  s32  data_offsets[data_total]; // NOTE: relative to signature
};

Mdb mdb @MDB_OFFSET;

DataAddress data_offset @MDB_OFFSET + mdb.data_offsets[0];

Bone bones[mdb.bone_total] @MDB_OFFSET + mdb.bone_offset;

Vertex vertices[data_offset.vertex_total] @MDB_OFFSET +
    mdb.data_offsets[0] + data_offset.position_offset;

Normal normals[data_offset.vertex_total] @MDB_OFFSET + 
    mdb.data_offsets[0] + data_offset.normal_offset;

Uv uvs[data_offset.vertex_total] @MDB_OFFSET + mdb.data_offsets[0] +
    data_offset.uv_offset;

Color colors[data_offset.vertex_total] @MDB_OFFSET + mdb.data_offsets[0] +
    data_offset.color_offset;

Weight weights[data_offset.vertex_total] @MDB_OFFSET +
    mdb.data_offsets[0] + data_offset.weight_offset;
